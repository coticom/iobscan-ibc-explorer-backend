/* eslint-disable @typescript-eslint/camelcase */
import * as mongoose from 'mongoose';
import {
    IbcTxType,
    IbcTxQueryType,
    AggregateResult24hr,
} from '../types/schemaTypes/ibc_tx.interface';
import {parseQuery} from '../helper/ibcTx.helper';
import {IbcTxStatus} from '../constant';

export const IbcTxSchema = new mongoose.Schema({
    record_id: String,
    sc_addr: String,
    dc_addr: String,
    sc_port: String,
    sc_channel: String,
    sc_chain_id: String,
    dc_port: String,
    dc_channel: String,
    dc_chain_id: String,
    sequence: String,
    status: Number,
    sc_tx_info: Object,
    dc_tx_info: Object,
    refunded_tx_info: Object,
    log: {
        sc_log: String,
        dc_log: String,
    },
    denoms: {
        sc_denom: String,
        dc_denom: String,
    },
    base_denom: String,
    create_at: {
        type: Number,
        default: Math.floor(new Date().getTime() / 1000),
    },
    update_at: {
        type: Number,
        default: Math.floor(new Date().getTime() / 1000),
    },
    tx_time: {
        type: Number,
        default: Math.floor(new Date().getTime() / 1000),
    },
});

IbcTxSchema.index({record_id: -1}, {unique: true});
IbcTxSchema.index({status: -1}, {background: true});
IbcTxSchema.index({status: -1, tx_time: -1, sc_chain_id: -1, 'denoms.sc_denom': -1}, {background: true});
IbcTxSchema.index({status: -1, tx_time: -1, dc_chain_id: -1, 'denoms.dc_denom': -1}, {background: true});

IbcTxSchema.statics = {
    async countActive(): Promise<number> {
        return this.count({
            tx_time: {$gte: Math.floor(new Date().getTime() / 1000) - 24 * 60 * 60},
            status: {
                $in: [
                    IbcTxStatus.SUCCESS,
                    IbcTxStatus.FAILED,
                    IbcTxStatus.PROCESSING,
                    IbcTxStatus.REFUNDED,
                ],
            },
        });
    },

    async countAll(): Promise<number> {
        return this.count({
            status: {
                $in: [
                    IbcTxStatus.SUCCESS,
                    IbcTxStatus.FAILED,
                    IbcTxStatus.PROCESSING,
                    IbcTxStatus.REFUNDED,
                ],
            },
        });
    },


    async findActiveChains24hr(dateNow: any): Promise<Array<AggregateResult24hr>> {

        return this.aggregate([
            {
                $match: {
                    tx_time: {$gte: dateNow - 24 * 60 * 60},
                    status: {
                        $in: [
                            IbcTxStatus.SUCCESS,
                            IbcTxStatus.FAILED,
                            IbcTxStatus.PROCESSING,
                            IbcTxStatus.REFUNDED,
                        ],
                    }
                }
            },
            {
                $group: {
                    _id: {sc_chain_id: "$sc_chain_id", dc_chain_id: "$dc_chain_id"}
                }
            }]);
    },

    async aggregateFindSrcChannels24hr(dateNow: any, chains: Array<string>): Promise<Array<AggregateResult24hr>> {

        return this.aggregate([
            {
                $match: {
                    sc_chain_id: {$in: chains},
                    tx_time: {$gte: dateNow - 24 * 60 * 60},
                    status: {
                        $in: [
                            IbcTxStatus.SUCCESS,
                            IbcTxStatus.FAILED,
                            IbcTxStatus.PROCESSING,
                            IbcTxStatus.REFUNDED,
                        ],
                    }
                }
            },
            {
                $group: {
                    _id: {sc_channel: "$sc_channel", sc_chain_id: "$sc_chain_id"}
                }
            }]);
    },

    async aggregateFindDesChannels24hr(dateNow: any, chains: Array<string>): Promise<Array<AggregateResult24hr>> {

        return this.aggregate([
            {
                $match: {
                    dc_chain_id: {$in: chains},
                    tx_time: {$gte: dateNow - 24 * 60 * 60},
                    status: {
                        $in: [
                            IbcTxStatus.SUCCESS,
                            IbcTxStatus.FAILED,
                            IbcTxStatus.PROCESSING,
                            IbcTxStatus.REFUNDED,
                        ],
                    }
                }
            },
            {
                $group: {
                    _id: {dc_channel: "$dc_channel", dc_chain_id: "$dc_chain_id"},
                }
            }]);
    },

    async countSuccess(): Promise<number> {
        return this.count({
            status: IbcTxStatus.SUCCESS,
        });
    },

    async countFaild(): Promise<number> {
        return this.count({
            status: {$in: [IbcTxStatus.FAILED, IbcTxStatus.REFUNDED]},
        });
    },

    async countTxList(query: IbcTxQueryType): Promise<number> {
        const queryParams = parseQuery(query);
        return this.count(queryParams);
    },

    async findTxList(query: IbcTxQueryType): Promise<IbcTxType[]> {
        const queryParams = parseQuery(query);
        const {page_num, page_size} = query;
        return this.find(queryParams, {_id: 0})
            .skip((Number(page_num) - 1) * Number(page_size))
            .limit(Number(page_size))
            .sort({tx_time: -1});
    },

    async findFirstTx(): Promise<IbcTxType> {
        return this.findOne().sort({tx_time: 1})
    },

    async queryTxList(query): Promise<IbcTxType[]> {
        const {status, limit} = query;
        return this.find({status}, {_id: 0})
            .sort({tx_time: 1})
            .limit(Number(limit));
    },

    // async distinctChainList(query): Promise<any> {
    //   const { type, dateNow, status } = query;
    //   return this.distinct(type, {
    //     update_at: { $gte: dateNow - 24 * 60 * 60 },
    //     status: { $in: status },
    //   });
    // },

    async updateIbcTx(ibcTx): Promise<void> {
        const {record_id} = ibcTx;
        const options = {upsert: true, new: false, setDefaultsOnInsert: true};
        return this.findOneAndUpdate({record_id}, ibcTx, options);
    },

    async insertManyIbcTx(ibcTx, cb): Promise<void> {
        return this.insertMany(ibcTx, cb);
    },
};
