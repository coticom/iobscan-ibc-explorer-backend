package repository

import (
	"context"
	"fmt"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/dto"
	"github.com/qiniu/qmgo/operator"

	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/entity"
	"github.com/qiniu/qmgo"
	opts "github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson"
	officialOpts "go.mongodb.org/mongo-driver/mongo/options"
)

type IChainFeeStatisticsRepo interface {
	CreateNew() error
	SwitchColl() error
	InsertMany(batch []*entity.IBCChainFeeStatistics) error
	InsertManyToNew(batch []*entity.IBCChainFeeStatistics) error
	BatchSwap(chain string, segmentStartTime, segmentEndTime int64, batch []*entity.IBCChainFeeStatistics) error
	ChainFeeStatistics(chain string, startTime, endTime int64) ([]*dto.DenomAmountStatisticsDTO, error)
	RelayerFeeStatistics(chain string, startTime, endTime int64) ([]*dto.DenomAmountStatisticsDTO, error)
}

var _ IChainFeeStatisticsRepo = new(ChainFeeStatisticsRepo)

type ChainFeeStatisticsRepo struct {
}

func (repo *ChainFeeStatisticsRepo) coll() *qmgo.Collection {
	return mgo.Database(ibcDatabase).Collection(entity.IBCChainFeeStatisticsCollName)
}

func (repo *ChainFeeStatisticsRepo) collNew() *qmgo.Collection {
	return mgo.Database(ibcDatabase).Collection(entity.IBCChainFeeStatisticsNewCollName)
}

func (repo *ChainFeeStatisticsRepo) CreateNew() error {
	ukOpts := officialOpts.Index().SetUnique(true).SetName("statistics_unique")
	uk := []string{"chain_name", "payer_type", "tx_status", "fee_denom", "segment_start_time", "segment_end_time"}
	if err := repo.collNew().CreateOneIndex(context.Background(), opts.IndexModel{Key: uk, IndexOptions: ukOpts}); err != nil {
		return err
	}

	indexOpts := officialOpts.Index()
	key := []string{"chain_name", "segment_start_time", "segment_end_time"}
	if err := repo.collNew().CreateOneIndex(context.Background(), opts.IndexModel{Key: key, IndexOptions: indexOpts}); err != nil {
		return err
	}

	return nil
}

func (repo *ChainFeeStatisticsRepo) SwitchColl() error {
	command := bson.D{{"renameCollection", fmt.Sprintf("%s.%s", ibcDatabase, entity.IBCChainFeeStatisticsNewCollName)},
		{"to", fmt.Sprintf("%s.%s", ibcDatabase, entity.IBCChainFeeStatisticsCollName)},
		{"dropTarget", true}}
	return mgo.Database(adminDatabase).RunCommand(context.Background(), command).Err()
}

func (repo *ChainFeeStatisticsRepo) InsertMany(batch []*entity.IBCChainFeeStatistics) error {
	if _, err := repo.coll().InsertMany(context.Background(), batch); err != nil {
		return err
	}
	return nil
}

func (repo *ChainFeeStatisticsRepo) InsertManyToNew(batch []*entity.IBCChainFeeStatistics) error {
	if _, err := repo.collNew().InsertMany(context.Background(), batch); err != nil {
		return err
	}
	return nil
}

func (repo *ChainFeeStatisticsRepo) BatchSwap(chain string, segmentStartTime, segmentEndTime int64, batch []*entity.IBCChainFeeStatistics) error {
	callback := func(sessCtx context.Context) (interface{}, error) {
		query := bson.M{
			"chain_name":         chain,
			"segment_start_time": segmentStartTime,
			"segment_end_time":   segmentEndTime,
		}
		if _, err := repo.coll().RemoveAll(sessCtx, query); err != nil {
			return nil, err
		}

		if len(batch) == 0 {
			return nil, nil
		}

		if _, err := repo.coll().InsertMany(sessCtx, batch); err != nil {
			return nil, err
		}

		return nil, nil
	}
	_, err := mgo.DoTransaction(context.Background(), callback)
	return err
}
func (repo *ChainFeeStatisticsRepo) ChainFeeStatistics(chain string, startTime, endTime int64) ([]*dto.DenomAmountStatisticsDTO, error) {
	match := bson.M{
		operator.Match: bson.M{
			"chain_name": chain,
			"segment_start_time": bson.M{
				operator.Gte: startTime,
			},
			"segment_end_time": bson.M{
				operator.Lte: endTime,
			},
		},
	}

	group := bson.M{
		operator.Group: bson.M{
			"_id": bson.M{
				"fee_denom": "$fee_denom",
			},
			"fee_amount": bson.M{
				operator.Sum: "$fee_amount",
			},
		},
	}

	project := bson.M{
		operator.Project: bson.M{
			"fee_denom":  "$_id.fee_denom",
			"fee_amount": "$fee_amount",
		},
	}

	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	var res []*dto.DenomAmountStatisticsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}

func (repo *ChainFeeStatisticsRepo) RelayerFeeStatistics(chain string, startTime, endTime int64) ([]*dto.DenomAmountStatisticsDTO, error) {
	match := bson.M{
		operator.Match: bson.M{
			"chain_name": chain,
			"payer_type": entity.RelayerPay,
			"segment_start_time": bson.M{
				operator.Gte: startTime,
			},
			"segment_end_time": bson.M{
				operator.Lte: endTime,
			},
		},
	}

	group := bson.M{
		operator.Group: bson.M{
			"_id": bson.M{
				"fee_denom": "$fee_denom",
			},
			"fee_amount": bson.M{
				operator.Sum: "$fee_amount",
			},
		},
	}

	project := bson.M{
		operator.Project: bson.M{
			"fee_denom":  "$_id.fee_denom",
			"fee_amount": "$fee_amount",
		},
	}

	var pipe []bson.M
	pipe = append(pipe, match, group, project)
	var res []*dto.DenomAmountStatisticsDTO
	err := repo.coll().Aggregate(context.Background(), pipe).All(&res)
	return res, err
}