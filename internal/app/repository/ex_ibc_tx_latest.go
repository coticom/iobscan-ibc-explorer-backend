package repository

import (
	"context"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/constant"
	"strings"

	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/dto"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/entity"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IExIbcTxRepo interface {
	InsertBatch(txs []*entity.ExIbcTx) error
	InsertBatchHistory(txs []*entity.ExIbcTx) error
	DeleteByRecordIds(recordIds []string) error
	FindAll(skip, limit int64) ([]*entity.ExIbcTx, error)
	FindByStatus(status []entity.IbcTxStatus, limit int64) ([]*entity.ExIbcTx, error)
	FindByTxTime(startTime, endTime, skip, limit int64) ([]*entity.ExIbcTx, error)
	FindHistoryByTxTime(startTime, endTime, skip, limit int64) ([]*entity.ExIbcTx, error)
	CountByStatus(status []entity.IbcTxStatus) (int64, error)
	FindAllHistory(skip, limit int64) ([]*entity.ExIbcTx, error)
	First() (*entity.ExIbcTx, error)
	FirstHistory() (*entity.ExIbcTx, error)
	Latest() (*entity.ExIbcTx, error)
	LatestHistory() (*entity.ExIbcTx, error)
	FindProcessingTxs(chainId string, limit int64) ([]*entity.ExIbcTx, error)
	FindProcessingHistoryTxs(chainId string, limit int64) ([]*entity.ExIbcTx, error)
	UpdateIbcTx(ibcTx *entity.ExIbcTx) error
	UpdateIbcHistoryTx(ibcTx *entity.ExIbcTx) error
	CountBaseDenomTransferTxs(startTime, endTime int64) ([]*dto.CountBaseDenomTxsDTO, error)
	CountBaseDenomHistoryTransferTxs(startTime, endTime int64) ([]*dto.CountBaseDenomTxsDTO, error)
	CountIBCTokenRecvTxs(startTime, endTime int64) ([]*dto.CountIBCTokenRecvTxsDTO, error)
	CountIBCTokenHistoryRecvTxs(startTime, endTime int64) ([]*dto.CountIBCTokenRecvTxsDTO, error)
	GetRelayerInfo(startTime, endTime int64) ([]*dto.GetRelayerInfoDTO, error)
	GetHistoryRelayerInfo(startTime, endTime int64) ([]*dto.GetRelayerInfoDTO, error)
	GetLatestTxTime() (int64, error)
	GetOneRelayerScTxPacketId(dto *dto.GetRelayerInfoDTO) (entity.ExIbcTx, error)
	GetHistoryOneRelayerScTxPacketId(dto *dto.GetRelayerInfoDTO) (entity.ExIbcTx, error)
	CountHistoryRelayerSuccessPacketTxs(startTime, endTime int64) ([]*dto.CountRelayerPacketTxsCntDTO, error)
	CountRelayerSuccessPacketTxs(startTime, endTime int64) ([]*dto.CountRelayerPacketTxsCntDTO, error)
	CountHistoryRelayerPacketAmount(startTime, endTime int64) ([]*dto.CountRelayerPacketAmountDTO, error)
	CountRelayerPacketTxsAndAmount(startTime, endTime int64) ([]*dto.CountRelayerPacketAmountDTO, error)
	AggrIBCChannelTxs(startTime, endTime int64) ([]*dto.AggrIBCChannelTxsDTO, error)
	AggrIBCChannelHistoryTxs(startTime, endTime int64) ([]*dto.AggrIBCChannelTxsDTO, error)
	Aggr24hActiveChannels(startTime int64) ([]*dto.Aggr24hActiveChannelsDTO, error)
	Aggr24hActiveChains(startTime int64) ([]*dto.Aggr24hActiveChainsDTO, error)
	Migrate(txs []*entity.ExIbcTx) error

	// special method
	UpdateDenomTrace(originRecordId string, ibcTx *entity.ExIbcTx) error
	UpdateDenomTraceHistory(originRecordId string, ibcTx *entity.ExIbcTx) error
	AddNewChainUpdate(scChainId, scChannel, dcChainId string) error
	AddNewChainUpdateHistory(scChainId, scChannel, dcChainId string) error
	AddNewChainUpdateFailedTx(scChainId, scChannel, dcChainId, dcChannel string) error
	AddNewChainUpdateHistoryFailedTx(scChainId, scChannel, dcChainId, dcChannel string) error
	UpdateBaseDenomInfo(baseDenom, baseDenomChainId, baseDenomNew, baseDenomChainIdNew string) error
	UpdateBaseDenomInfoHistory(baseDenom, baseDenomChainId, baseDenomNew, baseDenomChainIdNew string) error

	HistoryLatestCreateAt() (int64, error)
	HistoryCountAll(createAt int64, record bool) (int64, error)
	HistoryCountFailAll(createAt int64, record bool) (int64, error)
	HistoryCountSuccessAll(createAt int64, record bool) (int64, error)
	ActiveTxs24h(startTime int64) (int64, error)
	CountAll(stats []entity.IbcTxStatus) (int64, error)
	CountTransferTxs(query dto.IbcTxQuery) (int64, error)
	FindTransferTxs(query dto.IbcTxQuery, skip, limit int64) ([]*entity.ExIbcTx, error)
	TxDetail(hash string, history bool) ([]*entity.ExIbcTx, error)
	GetNeedAcknowledgeTxs(history bool) ([]*entity.ExIbcTx, error)
	GetNeedFailRecvPacketTxs(history bool) ([]*entity.ExIbcTx, error)
	UpdateOne(recordId string, history bool, data *entity.ExIbcTx) error

	// fix dc_chain_id
	FindDcChainIdEmptyTxs(startTime, endTime, skip, limit int64, isTargetHistory bool) ([]*entity.ExIbcTx, error)
	FixDcChainId(recordId, dcChainId, dcChannel string, originStatus entity.IbcTxStatus, isTargetHistory bool) error
	// fix base_denom_chain_id
	FindByBaseDenom(startTime, endTime int64, baseDenom, baseDenomChainId string, isTargetHistory bool) ([]*entity.ExIbcTx, error)
	UpdateBaseDenom(recordId, baseDenom, baseDenomChainId string, isTargetHistory bool) error
}

var _ IExIbcTxRepo = new(ExIbcTxRepo)

type ExIbcTxRepo struct {
}

func (repo *ExIbcTxRepo) coll() *qmgo.Collection {
	return mgo.Database(ibcDatabase).Collection(entity.ExIbcTx{}.CollectionName(false))
}

func (repo *ExIbcTxRepo) collHistory() *qmgo.Collection {
	return mgo.Database(ibcDatabase).Collection(entity.ExIbcTx{}.CollectionName(true))
}

func (repo *ExIbcTxRepo) InsertBatch(txs []*entity.ExIbcTx) error {
	_, err := repo.coll().InsertMany(context.Background(), txs, insertIgnoreErrOpt)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (repo *ExIbcTxRepo) InsertBatchHistory(txs []*entity.ExIbcTx) error {
	_, err := repo.collHistory().InsertMany(context.Background(), txs, insertIgnoreErrOpt)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (repo *ExIbcTxRepo) DeleteByRecordIds(recordIds []string) error {
	_, err := repo.coll().RemoveAll(context.Background(), bson.M{"record_id": bson.M{"$in": recordIds}})
	return err
}

func (repo *ExIbcTxRepo) FindAll(skip, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	err := repo.coll().Find(context.Background(), bson.M{}).Skip(skip).Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) FindAllHistory(skip, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	err := repo.collHistory().Find(context.Background(), bson.M{}).Skip(skip).Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) FindByStatus(status []entity.IbcTxStatus, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	err := repo.coll().Find(context.Background(), bson.M{"status": bson.M{"$in": status}}).Sort("tx_time").Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) FindByTxTime(startTime, endTime, skip, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	query := bson.M{
		"tx_time": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	err := repo.coll().Find(context.Background(), query).Sort("tx_time").Skip(skip).Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) FindHistoryByTxTime(startTime, endTime, skip, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	query := bson.M{
		"tx_time": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	err := repo.collHistory().Find(context.Background(), query).Sort("tx_time").Skip(skip).Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) CountByStatus(status []entity.IbcTxStatus) (int64, error) {
	var res int64
	res, err := repo.coll().Find(context.Background(), bson.M{"status": bson.M{"$in": status}}).Count()
	return res, err
}

func (repo *ExIbcTxRepo) First() (*entity.ExIbcTx, error) {
	var res entity.ExIbcTx
	err := repo.coll().Find(context.Background(), bson.M{}).Sort("create_at").One(&res)
	return &res, err
}

func (repo *ExIbcTxRepo) FirstHistory() (*entity.ExIbcTx, error) {
	var res entity.ExIbcTx
	err := repo.collHistory().Find(context.Background(), bson.M{}).Sort("create_at").One(&res)
	return &res, err
}

func (repo *ExIbcTxRepo) Latest() (*entity.ExIbcTx, error) {
	var res entity.ExIbcTx
	err := repo.coll().Find(context.Background(), bson.M{}).Sort("-create_at").One(&res)
	return &res, err
}

func (repo *ExIbcTxRepo) LatestHistory() (*entity.ExIbcTx, error) {
	var res entity.ExIbcTx
	err := repo.collHistory().Find(context.Background(), bson.M{}).Sort("-create_at").One(&res)
	return &res, err
}

func (repo *ExIbcTxRepo) FindProcessingTxs(chainId string, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	err := repo.coll().Find(context.Background(), bson.M{"sc_chain_id": chainId, "status": entity.IbcTxStatusProcessing}).Sort("next_try_time").Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) FindProcessingHistoryTxs(chainId string, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	err := repo.collHistory().Find(context.Background(), bson.M{"sc_chain_id": chainId, "status": entity.IbcTxStatusProcessing}).Sort("next_try_time").Limit(limit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) UpdateIbcTx(ibcTx *entity.ExIbcTx) error {
	return repo.coll().UpdateOne(context.Background(), bson.M{"record_id": ibcTx.RecordId}, bson.M{
		"$set": bson.M{
			"status":           ibcTx.Status,
			"denoms.dc_denom":  ibcTx.Denoms.DcDenom,
			"dc_tx_info":       ibcTx.DcTxInfo,
			"refunded_tx_info": ibcTx.RefundedTxInfo,
			"retry_times":      ibcTx.RetryTimes,
			"next_try_time":    ibcTx.NextTryTime,
			"process_info":     ibcTx.ProcessInfo,
			"update_at":        ibcTx.UpdateAt,
		},
	})
}

func (repo *ExIbcTxRepo) UpdateIbcHistoryTx(ibcTx *entity.ExIbcTx) error {
	return repo.collHistory().UpdateOne(context.Background(), bson.M{"record_id": ibcTx.RecordId}, bson.M{
		"$set": bson.M{
			"status":           ibcTx.Status,
			"denoms.dc_denom":  ibcTx.Denoms.DcDenom,
			"dc_tx_info":       ibcTx.DcTxInfo,
			"refunded_tx_info": ibcTx.RefundedTxInfo,
			"retry_times":      ibcTx.RetryTimes,
			"next_try_time":    ibcTx.NextTryTime,
			"process_info":     ibcTx.ProcessInfo,
			"update_at":        ibcTx.UpdateAt,
		},
	})
}

func (repo *ExIbcTxRepo) UpdateDenomTrace(originRecordId string, ibcTx *entity.ExIbcTx) error {
	return repo.coll().UpdateOne(context.Background(), bson.M{"record_id": originRecordId}, bson.M{
		"$set": bson.M{
			"base_denom_chain_id": ibcTx.BaseDenomChainId,
			"base_denom":          ibcTx.BaseDenom,
			"create_at":           ibcTx.CreateAt,
			"update_at":           ibcTx.UpdateAt,
		},
	})
}

func (repo *ExIbcTxRepo) UpdateDenomTraceHistory(originRecordId string, ibcTx *entity.ExIbcTx) error {
	return repo.collHistory().UpdateOne(context.Background(), bson.M{"record_id": originRecordId}, bson.M{
		"$set": bson.M{
			"record_id":           ibcTx.RecordId,
			"base_denom_chain_id": ibcTx.BaseDenomChainId,
			"base_denom":          ibcTx.BaseDenom,
			"create_at":           ibcTx.CreateAt,
			"update_at":           ibcTx.UpdateAt,
		},
	})
}

func (repo *ExIbcTxRepo) countBaseDenomTransferTxsPipe(startTime, endTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"create_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"status": bson.M{
				"$in": entity.IbcTxUsefulStatus,
			},
		},
	}

	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"base_denom":          "$base_denom",
				"base_denom_chain_id": "$base_denom_chain_id",
			},
			"count": bson.M{
				"$sum": 1,
			},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":                 0,
			"base_denom":          "$_id.base_denom",
			"base_denom_chain_id": "$_id.base_denom_chain_id",
			"count":               "$count",
		},
	}

	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) CountBaseDenomTransferTxs(startTime, endTime int64) ([]*dto.CountBaseDenomTxsDTO, error) {
	pipe := repo.countBaseDenomTransferTxsPipe(startTime, endTime)
	var res []*dto.CountBaseDenomTxsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) CountBaseDenomHistoryTransferTxs(startTime, endTime int64) ([]*dto.CountBaseDenomTxsDTO, error) {
	pipe := repo.countBaseDenomTransferTxsPipe(startTime, endTime)
	var res []*dto.CountBaseDenomTxsDTO
	err := repo.collHistory().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) countIBCTokenRecvTxsPipe(startTime, endTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"create_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"status": entity.IbcTxStatusSuccess,
		},
	}

	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"denom":    "$denoms.dc_denom",
				"chain_id": "$dc_chain_id",
			},
			"count": bson.M{
				"$sum": 1,
			},
		},
	}

	project :=
		bson.M{
			"$project": bson.M{
				"_id":      0,
				"denom":    "$_id.denom",
				"chain_id": "$_id.chain_id",
				"count":    "$count",
			}}

	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) CountIBCTokenRecvTxs(startTime, endTime int64) ([]*dto.CountIBCTokenRecvTxsDTO, error) {
	pipe := repo.countIBCTokenRecvTxsPipe(startTime, endTime)
	var res []*dto.CountIBCTokenRecvTxsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) CountIBCTokenHistoryRecvTxs(startTime, endTime int64) ([]*dto.CountIBCTokenRecvTxsDTO, error) {
	pipe := repo.countIBCTokenRecvTxsPipe(startTime, endTime)
	var res []*dto.CountIBCTokenRecvTxsDTO
	err := repo.collHistory().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) GetRelayerInfo(startTime, endTime int64) ([]*dto.GetRelayerInfoDTO, error) {
	pipe := repo.relayerInfoPipe(startTime, endTime)
	var res []*dto.GetRelayerInfoDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) GetHistoryRelayerInfo(startTime, endTime int64) ([]*dto.GetRelayerInfoDTO, error) {
	pipe := repo.relayerInfoPipe(startTime, endTime)
	var res []*dto.GetRelayerInfoDTO
	err := repo.collHistory().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) relayerInfoPipe(startTime, endTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"create_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"dc_tx_info.status": 1,
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"relayer":     "$dc_tx_info.msg.msg.signer",
				"sc_chain_id": "$sc_chain_id",
				"sc_channel":  "$sc_channel",
				"dc_chain_id": "$dc_chain_id",
				"dc_channel":  "$dc_channel",
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":              0,
			"dc_chain_address": "$_id.relayer",
			"sc_chain_id":      "$_id.sc_chain_id",
			"dc_chain_id":      "$_id.dc_chain_id",
			"sc_channel":       "$_id.sc_channel",
			"dc_channel":       "$_id.dc_channel",
		},
	}
	sort := bson.M{
		"$sort": bson.M{
			"tx_time": 1,
		},
	}
	var pipe []bson.M
	pipe = append(pipe, match, group, project, sort)
	return pipe

}

func (repo *ExIbcTxRepo) GetLatestTxTime() (int64, error) {
	var res *entity.ExIbcTx
	err := repo.coll().Find(context.Background(), bson.M{}).Select(bson.M{"tx_time": 1}).Sort("-tx_time").One(&res)
	if err != nil {
		return 0, err
	}
	return res.TxTime, nil
}

func (repo *ExIbcTxRepo) oneRelayerPacketCond(relayer *dto.GetRelayerInfoDTO) bson.M {
	return bson.M{
		"dc_tx_info.msg.msg.signer": relayer.DcChainAddress,
		"sc_chain_id":               relayer.ScChainId,
		"dc_chain_id":               relayer.DcChainId,
		"sc_channel":                relayer.ScChannel,
		"dc_channel":                relayer.DcChannel,
	}
}

func (repo *ExIbcTxRepo) GetOneRelayerScTxPacketId(dto *dto.GetRelayerInfoDTO) (entity.ExIbcTx, error) {
	var res entity.ExIbcTx
	err := repo.coll().Find(context.Background(), repo.oneRelayerPacketCond(dto)).
		Select(bson.M{"sc_tx_info.msg.msg.packet_id": 1}).Sort("-tx_time").Limit(1).One(&res)
	return res, err
}

func (repo *ExIbcTxRepo) GetHistoryOneRelayerScTxPacketId(dto *dto.GetRelayerInfoDTO) (entity.ExIbcTx, error) {
	var res entity.ExIbcTx
	err := repo.collHistory().Find(context.Background(), repo.oneRelayerPacketCond(dto)).
		Select(bson.M{"sc_tx_info.msg.msg.packet_id": 1}).Sort("-tx_time").Limit(1).One(&res)
	return res, err
}

func (repo *ExIbcTxRepo) relayerSuccessPacketCond(startTime, endTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"create_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"status": entity.IbcTxStatusSuccess,
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"dc_chain_id":         "$dc_chain_id",
				"dc_channel":          "$dc_channel",
				"sc_chain_id":         "$sc_chain_id",
				"sc_channel":          "$sc_channel",
				"relayer":             "$dc_tx_info.msg.msg.signer",
				"base_denom":          "$base_denom",
				"base_denom_chain_id": "$base_denom_chain_id",
			},
			"count": bson.M{
				"$sum": 1,
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":                 0,
			"dc_chain_address":    "$_id.relayer",
			"dc_chain_id":         "$_id.dc_chain_id",
			"dc_channel":          "$_id.dc_channel",
			"sc_chain_id":         "$_id.sc_chain_id",
			"sc_channel":          "$_id.sc_channel",
			"base_denom":          "$_id.base_denom",
			"base_denom_chain_id": "$_id.base_denom_chain_id",
			"count":               "$count",
		},
	}
	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) relayerPacketAmountCond(startTime, endTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"create_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"status": bson.M{
				"$in": entity.IbcTxUsefulStatus,
			},
			"sc_tx_info.status": entity.TxStatusSuccess,
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"dc_chain_id":         "$dc_chain_id",
				"dc_channel":          "$dc_channel",
				"sc_chain_id":         "$sc_chain_id",
				"sc_channel":          "$sc_channel",
				"relayer":             "$dc_tx_info.msg.msg.signer",
				"base_denom":          "$base_denom",
				"base_denom_chain_id": "$base_denom_chain_id",
			},
			"amount": bson.M{
				"$sum": bson.M{"$toDouble": "$sc_tx_info.msg_amount.amount"},
			},
			"count": bson.M{
				"$sum": 1,
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":                 0,
			"dc_chain_address":    "$_id.relayer",
			"dc_chain_id":         "$_id.dc_chain_id",
			"dc_channel":          "$_id.dc_channel",
			"sc_chain_id":         "$_id.sc_chain_id",
			"sc_channel":          "$_id.sc_channel",
			"base_denom":          "$_id.base_denom",
			"base_denom_chain_id": "$_id.base_denom_chain_id",
			"amount":              "$amount",
			"count":               "$count",
		},
	}
	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) CountHistoryRelayerSuccessPacketTxs(startTime, endTime int64) ([]*dto.CountRelayerPacketTxsCntDTO, error) {
	pipe := repo.relayerSuccessPacketCond(startTime, endTime)
	var res []*dto.CountRelayerPacketTxsCntDTO
	err := repo.collHistory().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) CountRelayerSuccessPacketTxs(startTime, endTime int64) ([]*dto.CountRelayerPacketTxsCntDTO, error) {
	pipe := repo.relayerSuccessPacketCond(startTime, endTime)
	var res []*dto.CountRelayerPacketTxsCntDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) CountRelayerPacketTxsAndAmount(startTime, endTime int64) ([]*dto.CountRelayerPacketAmountDTO, error) {
	pipe := repo.relayerPacketAmountCond(startTime, endTime)
	var res []*dto.CountRelayerPacketAmountDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) CountHistoryRelayerPacketAmount(startTime, endTime int64) ([]*dto.CountRelayerPacketAmountDTO, error) {
	pipe := repo.relayerPacketAmountCond(startTime, endTime)
	var res []*dto.CountRelayerPacketAmountDTO
	err := repo.collHistory().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) AggrIBCChannelTxsPipe(startTime, endTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"create_at": bson.M{
				"$gte": startTime,
				"$lte": endTime,
			},
			"status": bson.M{
				"$in": entity.IbcTxUsefulStatus,
			},
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"base_denom":          "$base_denom",
				"base_denom_chain_id": "$base_denom_chain_id",
				"sc_chain_id":         "$sc_chain_id",
				"dc_chain_id":         "$dc_chain_id",
				"sc_channel":          "$sc_channel",
				"dc_channel":          "$dc_channel",
			},
			"count": bson.M{
				"$sum": 1,
			},
			"amount": bson.M{
				"$sum": bson.M{
					"$toDouble": "$sc_tx_info.msg_amount.amount",
				},
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":                 0,
			"base_denom":          "$_id.base_denom",
			"base_denom_chain_id": "$_id.base_denom_chain_id",
			"sc_chain_id":         "$_id.sc_chain_id",
			"dc_chain_id":         "$_id.dc_chain_id",
			"sc_channel":          "$_id.sc_channel",
			"dc_channel":          "$_id.dc_channel",
			"count":               "$count",
			"amount":              "$amount",
		},
	}
	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) AggrIBCChannelTxs(startTime, endTime int64) ([]*dto.AggrIBCChannelTxsDTO, error) {
	pipe := repo.AggrIBCChannelTxsPipe(startTime, endTime)
	var res []*dto.AggrIBCChannelTxsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) AggrIBCChannelHistoryTxs(startTime, endTime int64) ([]*dto.AggrIBCChannelTxsDTO, error) {
	pipe := repo.AggrIBCChannelTxsPipe(startTime, endTime)
	var res []*dto.AggrIBCChannelTxsDTO
	err := repo.collHistory().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) aggr24hActiveChannelPipe(startTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"tx_time": bson.M{
				"$gte": startTime,
			},
			"$or": []bson.M{ //只统计成功、已退还状态、和第一段成功状态的但失败状态的
				{"status": bson.M{"$in": []entity.IbcTxStatus{entity.IbcTxStatusSuccess, entity.IbcTxStatusRefunded}}},
				{"status": entity.IbcTxStatusFailed, "sc_tx_info.status": entity.TxStatusSuccess},
			},
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"sc_chain_id": "$sc_chain_id",
				"dc_chain_id": "$dc_chain_id",
				"sc_channel":  "$sc_channel",
				"dc_channel":  "$dc_channel",
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":         0,
			"sc_chain_id": "$_id.sc_chain_id",
			"dc_chain_id": "$_id.dc_chain_id",
			"sc_channel":  "$_id.sc_channel",
			"dc_channel":  "$_id.dc_channel",
		},
	}
	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) aggr24hActiveChainsPipe(startTime int64) []bson.M {
	match := bson.M{
		"$match": bson.M{
			"tx_time": bson.M{
				"$gte": startTime,
			},
			"status": bson.M{
				"$in": entity.IbcTxUsefulStatus,
			},
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"sc_chain_id": "$sc_chain_id",
				"dc_chain_id": "$dc_chain_id",
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":         0,
			"sc_chain_id": "$_id.sc_chain_id",
			"dc_chain_id": "$_id.dc_chain_id",
		},
	}
	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	return pipe
}

func (repo *ExIbcTxRepo) Aggr24hActiveChannels(startTime int64) ([]*dto.Aggr24hActiveChannelsDTO, error) {
	pipe := repo.aggr24hActiveChannelPipe(startTime)
	var res []*dto.Aggr24hActiveChannelsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) Aggr24hActiveChains(startTime int64) ([]*dto.Aggr24hActiveChainsDTO, error) {
	pipe := repo.aggr24hActiveChainsPipe(startTime)
	var res []*dto.Aggr24hActiveChainsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) Migrate(txs []*entity.ExIbcTx) error {
	if len(txs) == 0 {
		return nil
	}

	var recordIds []string
	for _, v := range txs {
		recordIds = append(recordIds, v.RecordId)
	}

	callback := func(sessCtx context.Context) (interface{}, error) {
		if err := repo.InsertBatchHistory(txs); err != nil {
			return nil, err
		}

		if err := repo.DeleteByRecordIds(recordIds); err != nil {
			return nil, err
		}

		return nil, nil
	}
	_, err := mgo.DoTransaction(context.Background(), callback)
	return err
}

func (repo *ExIbcTxRepo) AddNewChainUpdate(scChainId, scChannel, dcChainId string) error {
	query := bson.M{
		"sc_chain_id": scChainId,
		"sc_channel":  scChannel,
		"status": bson.M{
			"$in": []entity.IbcTxStatus{entity.IbcTxStatusSetting, entity.IbcTxStatusProcessing},
		},
	}

	_, err := repo.coll().UpdateAll(context.Background(), query, bson.M{
		"$set": bson.M{
			"dc_chain_id": dcChainId,
			"status":      entity.IbcTxStatusProcessing,
		},
	})
	return err
}

func (repo *ExIbcTxRepo) AddNewChainUpdateHistory(scChainId, scChannel, dcChainId string) error {
	query := bson.M{
		"sc_chain_id": scChainId,
		"sc_channel":  scChannel,
		"status": bson.M{
			"$in": []entity.IbcTxStatus{entity.IbcTxStatusSetting, entity.IbcTxStatusProcessing},
		},
	}

	_, err := repo.collHistory().UpdateAll(context.Background(), query, bson.M{
		"$set": bson.M{
			"dc_chain_id": dcChainId,
			"status":      entity.IbcTxStatusProcessing,
		},
	})
	return err
}

func (repo *ExIbcTxRepo) AddNewChainUpdateFailedTx(scChainId, scChannel, dcChainId, dcChannel string) error {
	query := bson.M{
		"sc_chain_id": scChainId,
		"sc_channel":  scChannel,
		"status":      entity.IbcTxStatusFailed,
	}

	_, err := repo.coll().UpdateAll(context.Background(), query, bson.M{
		"$set": bson.M{
			"dc_chain_id": dcChainId,
			"dc_channel":  dcChannel,
		},
	})
	return err
}

func (repo *ExIbcTxRepo) AddNewChainUpdateHistoryFailedTx(scChainId, scChannel, dcChainId, dcChannel string) error {
	query := bson.M{
		"sc_chain_id": scChainId,
		"sc_channel":  scChannel,
		"status":      entity.IbcTxStatusFailed,
	}

	_, err := repo.collHistory().UpdateAll(context.Background(), query, bson.M{
		"$set": bson.M{
			"dc_chain_id": dcChainId,
			"dc_channel":  dcChannel,
		},
	})
	return err
}

func (repo *ExIbcTxRepo) UpdateBaseDenomInfo(baseDenom, baseDenomChainId, baseDenomNew, baseDenomChainIdNew string) error {
	query := bson.M{
		"base_denom":          baseDenom,
		"base_denom_chain_id": baseDenomChainId,
	}
	_, err := repo.coll().UpdateAll(context.Background(), query, bson.M{
		"$set": bson.M{
			"base_denom":          baseDenomNew,
			"base_denom_chain_id": baseDenomChainIdNew,
		},
	})
	return err
}

func (repo *ExIbcTxRepo) UpdateBaseDenomInfoHistory(baseDenom, baseDenomChainId, baseDenomNew, baseDenomChainIdNew string) error {
	query := bson.M{
		"base_denom":          baseDenom,
		"base_denom_chain_id": baseDenomChainId,
	}
	_, err := repo.collHistory().UpdateAll(context.Background(), query, bson.M{
		"$set": bson.M{
			"base_denom":          baseDenomNew,
			"base_denom_chain_id": baseDenomChainIdNew,
		},
	})
	return err
}

func (repo *ExIbcTxRepo) HistoryLatestCreateAt() (int64, error) {
	var res entity.ExIbcTx
	err := repo.collHistory().Find(context.Background(), bson.M{}).Sort("-create_at").One(&res)
	if err != nil {
		return 0, err
	}
	return res.CreateAt, nil
}

func (repo *ExIbcTxRepo) HistoryCountAll(createAt int64, record bool) (int64, error) {
	query := bson.M{
		"create_at": bson.M{
			"$gte": createAt,
		},
		"status": bson.M{
			"$in": entity.IbcTxUsefulStatus,
		},
	}
	//记录create_at时间点统计的数量
	if record {
		query = bson.M{
			"create_at": createAt,
			"status": bson.M{
				"$in": entity.IbcTxUsefulStatus,
			},
		}
	}
	return repo.collHistory().Find(context.Background(), query).Count()
}

func (repo *ExIbcTxRepo) HistoryCountFailAll(createAt int64, record bool) (int64, error) {
	query := bson.M{
		"create_at": bson.M{
			"$gte": createAt,
		},
		"status": bson.M{"$in": []entity.IbcTxStatus{entity.IbcTxStatusFailed, entity.IbcTxStatusRefunded}},
	}
	//记录create_at时间点统计的数量
	if record {
		query = bson.M{
			"create_at": createAt,
			"status":    bson.M{"$in": []entity.IbcTxStatus{entity.IbcTxStatusFailed, entity.IbcTxStatusRefunded}},
		}
	}
	return repo.collHistory().Find(context.Background(), query).Count()
}

func (repo *ExIbcTxRepo) HistoryCountSuccessAll(createAt int64, record bool) (int64, error) {
	query := bson.M{
		"create_at": bson.M{
			"$gte": createAt,
		},
		"status": entity.IbcTxStatusSuccess,
	}
	//记录create_at时间点统计的数量
	if record {
		query = bson.M{
			"create_at": createAt,
			"status":    entity.IbcTxStatusSuccess,
		}
	}
	return repo.collHistory().Find(context.Background(), query).Count()
}

func (repo *ExIbcTxRepo) ActiveTxs24h(startTime int64) (int64, error) {
	query := bson.M{
		"tx_time": bson.M{
			"$gte": startTime,
		},
		"status": bson.M{
			"$in": entity.IbcTxUsefulStatus,
		},
	}
	return repo.coll().Find(context.Background(), query).Count()
}

func (repo *ExIbcTxRepo) CountAll(stats []entity.IbcTxStatus) (int64, error) {
	query := bson.M{
		"status": bson.M{
			"$in": stats,
		},
	}
	return repo.coll().Find(context.Background(), query).Count()
}

func parseQuery(queryCond dto.IbcTxQuery) bson.M {
	query := bson.M{}

	//time
	if queryCond.StartTime > 0 && queryCond.EndTime > 0 {
		query["tx_time"] = bson.M{
			"$gte": queryCond.StartTime,
			"$lte": queryCond.EndTime,
		}
	} else if queryCond.StartTime > 0 {
		query["tx_time"] = bson.M{
			"$gte": queryCond.StartTime,
		}
	} else if queryCond.EndTime > 0 {
		query["tx_time"] = bson.M{
			"$lte": queryCond.EndTime,
		}
	}
	//chain
	if length := len(queryCond.ChainId); length > 0 {
		switch length {
		case 1:
			// transfer_chain or recv_chain
			if queryCond.ChainId[0] != constant.AllChain {
				query["$or"] = []bson.M{
					{"sc_chain_id": queryCond.ChainId[0]},
					{"dc_chain_id": queryCond.ChainId[0]},
				}
			}
		case 2:
			//transfer_chain and recv_chain
			if queryCond.ChainId[0] == queryCond.ChainId[1] && queryCond.ChainId[0] == constant.AllChain {
				// nothing to do
			} else {
				value := strings.Join(queryCond.ChainId, ",")
				if strings.Contains(value, constant.AllChain) {
					index := strings.Index(value, constant.AllChain)
					if index > 0 { //chain-id,allchain
						query["sc_chain_id"] = queryCond.ChainId[0]
					} else { //allchain,chain-id
						query["dc_chain_id"] = queryCond.ChainId[1]
					}

				} else {
					query["$and"] = []bson.M{
						{"sc_chain_id": queryCond.ChainId[0]},
						{"dc_chain_id": queryCond.ChainId[1]},
					}
				}
			}

		}
	}
	//token
	if len(queryCond.Token) > 0 {
		if strings.HasPrefix(queryCond.Token[0], "ibc/") {
			cond := []bson.M{
				{"denoms.sc_denom": queryCond.Token[0]},
				{"denoms.dc_denom": queryCond.Token[0]},
			}
			if _, exist := query["$or"]; exist {
				query["$and"] = []bson.M{
					{"$or": cond},
				}
			} else {
				query["$or"] = cond
			}
		} else {
			query["base_denom"] = bson.M{
				"$in": queryCond.Token,
			}
		}
	}
	//origin_chain_id
	if len(queryCond.BaseDenomChainId) > 0 {
		query["base_denom_chain_id"] = queryCond.BaseDenomChainId
	}

	//status
	if len(queryCond.Status) == 0 {
		query["status"] = bson.M{
			"$in": entity.IbcTxUsefulStatus,
		}
	} else {
		query["status"] = bson.M{
			"$in": queryCond.Status,
		}
	}
	return query
}

func (repo *ExIbcTxRepo) CountTransferTxs(query dto.IbcTxQuery) (int64, error) {
	return repo.coll().Find(context.Background(), parseQuery(query)).Count()
}

func (repo *ExIbcTxRepo) FindTransferTxs(query dto.IbcTxQuery, skip, limit int64) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	err := repo.coll().Find(context.Background(), parseQuery(query)).Skip(skip).Limit(limit).Sort("-tx_time").All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) TxDetail(hash string, history bool) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	query := bson.M{
		"status": bson.M{
			"$in": entity.IbcTxUsefulStatus,
		},
		"$or": []bson.M{
			{"sc_tx_info.hash": hash},
			{"dc_tx_info.hash": hash},
		},
	}
	if history {
		err := repo.collHistory().Find(context.Background(), query).All(&res)
		return res, err
	}
	err := repo.coll().Find(context.Background(), query).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) GetNeedAcknowledgeTxs(history bool) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	//查询"成功"和"已退还"状态的没有refunded_tx_info的数据
	query := bson.M{
		"status": bson.M{
			"$in": []entity.IbcTxStatus{entity.IbcTxStatusSuccess, entity.IbcTxStatusRefunded},
		},
		"refunded_tx_info.status": bson.M{"$exists": false},
	}
	if history {
		err := repo.collHistory().Find(context.Background(), query).Limit(constant.DefaultLimit).All(&res)
		return res, err
	}
	err := repo.coll().Find(context.Background(), query).Limit(constant.DefaultLimit).All(&res)
	return res, err
}

func (repo *ExIbcTxRepo) UpdateOne(recordId string, history bool, data *entity.ExIbcTx) error {
	if history {
		err := repo.collHistory().UpdateOne(context.Background(), bson.M{
			"record_id": recordId,
		}, data)
		return err
	}
	err := repo.coll().UpdateOne(context.Background(), bson.M{
		"record_id": recordId,
	}, data)
	return err
}

func (repo *ExIbcTxRepo) FindDcChainIdEmptyTxs(startTime, endTime, skip, limit int64, isTargetHistory bool) ([]*entity.ExIbcTx, error) {
	query := bson.M{
		"create_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"status": bson.M{
			"$in": []entity.IbcTxStatus{entity.IbcTxStatusProcessing, entity.IbcTxStatusSetting},
		},
		"dc_chain_id": "",
	}

	var txs []*entity.ExIbcTx
	var err error
	if isTargetHistory {
		err = repo.collHistory().Find(context.Background(), query).Skip(skip).Sort("-create_at").Limit(limit).All(&txs)
	} else {
		err = repo.coll().Find(context.Background(), query).Skip(skip).Sort("-create_at").Limit(limit).All(&txs)
	}
	return txs, err
}

func (repo *ExIbcTxRepo) FixDcChainId(recordId, dcChainId, dcChannel string, originStatus entity.IbcTxStatus, isTargetHistory bool) error {
	set := bson.M{}
	if dcChainId == "" {
		if originStatus == entity.IbcTxStatusProcessing {
			set["status"] = entity.IbcTxStatusSetting
		} else {
			return nil
		}
	} else {
		set["dc_chain_id"] = dcChainId
		set["dc_channel"] = dcChannel
		if originStatus == entity.IbcTxStatusSetting {
			set["status"] = entity.IbcTxStatusProcessing
		}
	}

	update := bson.M{
		"$set": set,
	}
	filter := bson.M{
		"record_id": recordId,
	}
	if isTargetHistory {
		return repo.collHistory().UpdateOne(context.Background(), filter, update)
	}
	return repo.coll().UpdateOne(context.Background(), filter, update)
}

func (repo *ExIbcTxRepo) FindByBaseDenom(startTime, endTime int64, baseDenom, baseDenomChainId string, isTargetHistory bool) ([]*entity.ExIbcTx, error) {
	query := bson.M{
		"create_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
		"base_denom":          baseDenom,
		"base_denom_chain_id": baseDenomChainId,
	}

	var txs []*entity.ExIbcTx
	var err error
	if isTargetHistory {
		err = repo.collHistory().Find(context.Background(), query).All(&txs)
	} else {
		err = repo.coll().Find(context.Background(), query).All(&txs)
	}

	return txs, err
}

func (repo *ExIbcTxRepo) UpdateBaseDenom(recordId, baseDenom, baseDenomChainId string, isTargetHistory bool) error {
	update := bson.M{
		"$set": bson.M{
			"base_denom":          baseDenom,
			"base_denom_chain_id": baseDenomChainId,
		}}
	if isTargetHistory {
		return repo.collHistory().UpdateOne(context.Background(), bson.M{"record_id": recordId}, update)
	}
	return repo.coll().UpdateOne(context.Background(), bson.M{"record_id": recordId}, update)
}

func (repo *ExIbcTxRepo) GetNeedFailRecvPacketTxs(history bool) ([]*entity.ExIbcTx, error) {
	var res []*entity.ExIbcTx
	//查询"已退还"状态的没有dc_tx_info的数据
	query := bson.M{
		"status":            entity.IbcTxStatusRefunded,
		"dc_tx_info.status": bson.M{"$exists": false},
	}
	if history {
		err := repo.collHistory().Find(context.Background(), query).Limit(constant.DefaultLimit).All(&res)
		return res, err
	}
	err := repo.coll().Find(context.Background(), query).Limit(constant.DefaultLimit).All(&res)
	return res, err
}
