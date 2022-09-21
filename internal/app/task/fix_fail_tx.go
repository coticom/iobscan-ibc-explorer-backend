package task

import (
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/constant"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/entity"
	"github.com/qiniu/qmgo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"sync"
)

type FixFailTxTask struct {
}

var _ OneOffTask = new(FixFailTxTask)

func (t *FixFailTxTask) Name() string {
	return "fix_fail_tx_task"
}

func (t *FixFailTxTask) Switch() bool {
	return false
}

func (t *FixFailTxTask) Run() int {
	segments, err := getSegment(segmentStepLatest)
	if err != nil {
		logrus.Errorf("task %s getSegment error, %v", t.Name(), err)
		return -1
	}

	historySegments, err := getHistorySegment(segmentStepHistory)
	if err != nil {
		logrus.Errorf("task %s getHistorySegment error, %v", t.Name(), err)
		return -1
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := t.fixFailTxs(ibcTxTargetLatest, segments)
		logrus.Infof("task %s fix latest end, %v", t.Name(), err)
	}()

	go func() {
		defer wg.Done()
		err := t.fixFailTxs(ibcTxTargetHistory, historySegments)
		logrus.Infof("task %s fix history end, %v", t.Name(), err)
	}()

	wg.Wait()
	return 1
}

func (t *FixFailTxTask) FixAcknowledgeTx(recordId string, ackTx *entity.Tx, history bool, status entity.IbcTxStatus) error {
	refundedTxInfo := &entity.TxInfo{
		Hash:      ackTx.TxHash,
		Height:    ackTx.Height,
		Time:      ackTx.Time,
		Status:    ackTx.Status,
		Fee:       ackTx.Fee,
		Memo:      ackTx.Memo,
		Signers:   ackTx.Signers,
		MsgAmount: nil,
		Msg:       getMsgByType(*ackTx, constant.MsgTypeAcknowledgement),
	}
	update := bson.M{
		"refunded_tx_info": refundedTxInfo,
		"status":           status,
	}
	return ibcTxRepo.UpdateOne(recordId, history, bson.M{
		"$set": update,
	})
}

func (t *FixFailTxTask) FixRecvPacketTxs(recordId string, recvTx, ackTx *entity.Tx, history bool, status entity.IbcTxStatus) error {
	if status <= 0 {
		return nil
	}
	update := bson.M{
		"status": status,
	}
	if recvTx != nil {
		dcTxInfo := &entity.TxInfo{
			Hash:      recvTx.TxHash,
			Height:    recvTx.Height,
			Time:      recvTx.Time,
			Status:    recvTx.Status,
			Fee:       recvTx.Fee,
			Memo:      recvTx.Memo,
			Signers:   recvTx.Signers,
			MsgAmount: nil,
			Msg:       getMsgByType(*recvTx, constant.MsgTypeRecvPacket),
		}
		update["dc_tx_info"] = dcTxInfo
	} else if status == entity.IbcTxStatusProcessing {
		//"处理中"将第二段数据清空
		update["dc_tx_info"] = bson.M{}
	}

	if ackTx != nil {
		refundedTxInfo := &entity.TxInfo{
			Hash:      ackTx.TxHash,
			Height:    ackTx.Height,
			Time:      ackTx.Time,
			Status:    ackTx.Status,
			Fee:       ackTx.Fee,
			Memo:      ackTx.Memo,
			Signers:   ackTx.Signers,
			MsgAmount: nil,
			Msg:       getMsgByType(*ackTx, constant.MsgTypeAcknowledgement),
		}
		update["refunded_tx_info"] = refundedTxInfo
	}

	return ibcTxRepo.UpdateOne(recordId, history, bson.M{
		"$set": update,
	})
}

//检查recv_packet的events中是否包含write_acknowledgement
func (t *FixFailTxTask) checkWriteAcknowledgeError(tx *entity.Tx) (writeAckOk, findWriteAck bool, ackRes string) {
	for msgIndex, msg := range tx.DocTxMsgs {
		if msg.Type != constant.MsgTypeRecvPacket {
			continue
		}
		for _, event := range tx.EventsNew {
			if event.MsgIndex != uint32(msgIndex) {
				continue
			}

			for _, ee := range event.Events {
				if ee.Type == "write_acknowledgement" {
					findWriteAck = true
					for _, attr := range ee.Attributes {
						if attr.Key == "packet_ack" {
							writeAckOk = !strings.Contains(attr.Value, "error")
							ackRes = attr.Value
							break
						}
					}
					break
				}
			}
		}
	}
	return
}

func (t *FixFailTxTask) fixFailTxs(target string, segments []*segment) error {
	const limit int64 = 1000
	isTargetHistory := false
	if target == ibcTxTargetHistory {
		isTargetHistory = true
	}

	for _, v := range segments {
		logrus.Infof("task %s fix %s %d-%d", t.Name(), target, v.StartTime, v.EndTime)
		var skip int64 = 0
		for {
			txs, err := ibcTxRepo.FindFailStatusTxs(v.StartTime, v.EndTime, skip, limit, isTargetHistory)
			if err != nil {
				logrus.Errorf("task %s FindFailToRefundStatusTxs %s %d-%d err, %v", t.Name(), target, v.StartTime, v.EndTime, err)
				return err
			}

			for _, val := range txs {
				bindedTx, err := txRepo.GetTxByHash(val.DcChainId, val.DcTxInfo.Hash)
				if err != nil {
					logrus.Errorf("task %s  %s err, chain_id: %s, packet_id: %s, %v", t.Name(), target, val.ScChainId, val.ScTxInfo.Msg.CommonMsg().PacketId, err)
					return err
				}
				wAckOk, findWriteAck, ackRes := t.checkWriteAcknowledgeError(&bindedTx)
				if findWriteAck { //关联的recv_packet有ack，根据ack找acknowledge tx
					ackTx, err := findAckTx(val, ackRes)
					if err != nil {
						return err
					}
					if ackTx != nil {
						var status entity.IbcTxStatus
						if wAckOk { //status: fail->success
							status = entity.IbcTxStatusSuccess
						} else { //status: fail->refund
							status = entity.IbcTxStatusRefunded
						}
						err = t.FixAcknowledgeTx(val.RecordId, ackTx, isTargetHistory, status)
						if err != nil && err != qmgo.ErrNoSuchDocuments {
							logrus.Errorf("task %s  %s err, chain_id: %s, packet_id: %s, %v", t.Name(), target, val.ScChainId, val.ScTxInfo.Msg.CommonMsg().PacketId, err)
							return err
						}
					} else {
						logrus.Debugf("status:%d recv_packet(chain_id:%s hash:%s) findWriteAck is ok,but no found acknowledge tx(chain_id:%s) tx",
							val.Status, val.DcChainId, bindedTx.TxHash, val.ScChainId)
						//status:fail->process
						err = t.FixRecvPacketTxs(val.RecordId, nil, nil, isTargetHistory, entity.IbcTxStatusProcessing)
						if err != nil && err != qmgo.ErrNoSuchDocuments {
							logrus.Errorf("task %s FixRecvPacketTxs %s err, chain_id: %s, packet_id: %s, %v",
								t.Name(), target, val.ScChainId, val.ScTxInfo.Msg.CommonMsg().PacketId, err)
							return err
						}
					}
				} else {
					//status: fail->success or fail->refund or fail->process
					recvTxs, err := txRepo.GetRecvPacketTxs(val.DcChainId, val.ScTxInfo.Msg.CommonMsg().PacketId)
					if err != nil {
						return err
					}

					var (
						recvTx, ackTx          *entity.Tx
						varAckRes              string
						ackOk, varfindWriteAck bool
						status                 entity.IbcTxStatus
					)

					for _, recvOne := range recvTxs {
						ackOk, varfindWriteAck, varAckRes = t.checkWriteAcknowledgeError(recvOne)
						if varfindWriteAck {
							recvTx = recvOne
							break
						}
					}

					if varfindWriteAck {
						ackTx, err = findAckTx(val, varAckRes)
						if err != nil {
							return err
						}

						if ackOk {
							status = entity.IbcTxStatusSuccess
						} else {
							status = entity.IbcTxStatusRefunded
						}
					} else { //没有找到包含writeAck的recv_packet
						status = entity.IbcTxStatusProcessing
					}
					err = t.FixRecvPacketTxs(val.RecordId, recvTx, ackTx, isTargetHistory, status)
					if err != nil && err != qmgo.ErrNoSuchDocuments {
						logrus.Errorf("task %s FixRecvPacketTxs %s err, chain_id: %s, packet_id: %s, %v", t.Name(), target, val.ScChainId, val.ScTxInfo.Msg.CommonMsg().PacketId, err)
						return err
					}
				}

			}

			if int64(len(txs)) < limit {
				break
			}
			skip += limit
		}
	}
	return nil
}

func findAckTx(val *entity.ExIbcTx, ackRes string) (*entity.Tx, error) {
	var ackTx *entity.Tx
	ackTxs, err := txRepo.GetAcknowledgeTxs(val.ScChainId, val.ScTxInfo.Msg.CommonMsg().PacketId)
	if err != nil {
		return nil, err
	}
	if len(ackTxs) > 0 {
		for _, ackOne := range ackTxs {
			for _, msg := range ackOne.DocTxMsgs {
				if msg.Type == constant.MsgTypeAcknowledgement {
					//ack tx的msg.ack 与 recv_packet的events ack一样
					if msg.AckPacketMsg().Acknowledgement == ackRes {
						ackTx = ackOne
						break
					}
				}
			}
		}

	}
	return ackTx, nil
}