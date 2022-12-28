package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/constant"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/errors"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/entity"
	"github.com/bianjieai/iobscan-ibc-explorer-backend/internal/app/model/vo"
	v8 "github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

type IOverviewService interface {
	MarketHeatmap() (*vo.MarketHeatmapResp, errors.Error)
	TokenDistribution(req *vo.TokenDistributionReq) (*vo.TokenDistributionResp, errors.Error)
}

var _ IOverviewService = new(OverviewService)

type OverviewService struct {
}

func (svc *OverviewService) MarketHeatmap() (*vo.MarketHeatmapResp, errors.Error) {
	nowTime := time.Now()
	before24hTime := nowTime.AddDate(0, 0, -1)
	statisticsTime, err := denomHeatmapRepo.FindLastStatisticsTime(nowTime)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	denomHeatmapList, err := denomHeatmapRepo.FindByStatisticsTime(statisticsTime)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	resp := svc.buildMarketHeatmapResp(denomHeatmapList)

	before24hStatisticsTime, err := denomHeatmapRepo.FindLastStatisticsTime(before24hTime)
	if err != nil {
		return resp, nil
	}

	before24hDenomHeatmapList, err := denomHeatmapRepo.FindByStatisticsTime(before24hStatisticsTime)
	if err != nil {
		return resp, nil
	}

	if len(before24hDenomHeatmapList) > 0 {
		resp = svc.completeMarketHeatmapResp(resp, before24hDenomHeatmapList)
	}

	return resp, nil
}

func (svc *OverviewService) completeMarketHeatmapResp(resp *vo.MarketHeatmapResp, denomHeatmapList []*entity.DenomHeatmap) *vo.MarketHeatmapResp {
	oldPriceMap := make(map[string]float64, len(denomHeatmapList))
	oldTotalMarketMap := decimal.Zero
	for _, v := range denomHeatmapList {
		oldPriceMap[fmt.Sprintf("%s_%s", v.Chain, v.Denom)] = v.Price
		marketCapDecimal, _ := decimal.NewFromString(v.MarketCap)
		oldTotalMarketMap = oldTotalMarketMap.Add(marketCapDecimal)
	}

	growthRateFunc := func(d1, d2 decimal.Decimal) (rate float64, trend string) {
		rate = d1.DivRound(d2, 4).
			Sub(decimal.NewFromInt(1)).InexactFloat64()
		if rate >= 0 {
			trend = constant.IncreaseSymbol
		} else {
			trend = constant.DecreaseSymbol
		}

		rate = math.Abs(rate)
		return rate, trend
	}

	for i, v := range resp.Items {
		if price, ok := oldPriceMap[fmt.Sprintf("%s_%s", v.Chain, v.Denom)]; ok {
			if price == 0 {
				continue
			}

			resp.Items[i].PriceGrowthRate, resp.Items[i].PriceTrend = growthRateFunc(decimal.NewFromFloat(v.Price), decimal.NewFromFloat(price))
		}
	}

	if !oldTotalMarketMap.Equal(decimal.Zero) {
		totalMarketCap, _ := decimal.NewFromString(resp.TotalInfo.TotalMarketCap)
		resp.TotalInfo.MarketCapGrowthRate, resp.TotalInfo.MarketCapTrend = growthRateFunc(totalMarketCap, oldTotalMarketMap)
	}

	return resp
}

func (svc *OverviewService) buildMarketHeatmapResp(denomHeatmapList []*entity.DenomHeatmap) *vo.MarketHeatmapResp {
	var stableCoinMap entity.IBCBaseDenomMap
	stableCoins, err := authDenomRepo.FindStableCoins()
	if err == nil {
		stableCoinMap = stableCoins.ConvertToMap()
	}

	heatmapItemList := make([]vo.HeatmapItem, 0, len(denomHeatmapList))
	totalMarketCap := decimal.Zero
	stablecoinsMarketCap := decimal.Zero
	transferVolumeTotal := decimal.Zero
	var atomPrice float64
	atomMarketCap := decimal.Zero

	for _, v := range denomHeatmapList {
		marketCapDecimal, _ := decimal.NewFromString(v.MarketCap)
		totalMarketCap = totalMarketCap.Add(marketCapDecimal)

		transferVolume24hDecimal, _ := decimal.NewFromString(v.TransferVolume24h)
		transferVolumeTotal = transferVolumeTotal.Add(transferVolume24hDecimal)

		if _, ok := stableCoinMap[fmt.Sprintf("%s%s", v.Chain, v.Denom)]; ok {
			stablecoinsMarketCap = stablecoinsMarketCap.Add(marketCapDecimal)
		}

		if v.Chain == constant.ChainNameCosmosHub && v.Denom == constant.DenomAtom {
			atomPrice = v.Price
			atomMarketCap, _ = decimal.NewFromString(v.MarketCap)
		}

		heatmapItemList = append(heatmapItemList, vo.HeatmapItem{
			Price:               v.Price,
			PriceGrowthRate:     0,
			PriceTrend:          constant.IncreaseSymbol,
			Denom:               v.Denom,
			Chain:               v.Chain,
			MarketCapValue:      v.MarketCap,
			TransferVolumeValue: v.TransferVolume24h,
		})

	}

	var atomDominance float64
	if !totalMarketCap.Equal(decimal.Zero) {
		atomDominance = atomMarketCap.DivRound(totalMarketCap, 4).InexactFloat64()
	}

	heatmapTotalInfo := vo.HeatmapTotalInfo{
		StablecoinsMarketCap: stablecoinsMarketCap.String(),
		TotalMarketCap:       totalMarketCap.String(),
		MarketCapGrowthRate:  0,
		MarketCapTrend:       constant.IncreaseSymbol,
		TransferVolumeTotal:  transferVolumeTotal.String(),
		AtomPrice:            atomPrice,
		AtomDominance:        atomDominance,
	}

	return &vo.MarketHeatmapResp{
		Items:     heatmapItemList,
		TotalInfo: heatmapTotalInfo,
	}
}

func (svc *OverviewService) TokenDistribution(req *vo.TokenDistributionReq) (*vo.TokenDistributionResp, errors.Error) {
	ibcDenoms, err := denomRepo.FindByBaseDenom(req.BaseDenom, req.BaseDenomChain)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	getHops := func(denomPath string) int {
		return strings.Count(denomPath, "/channel")
	}
	mapHopsData := make(map[int]entity.IBCDenomList, 1)
	mapChainData := make(map[string]string, 1)
	for _, val := range ibcDenoms {

		amount, err := supportCache.GetSupply(val.Chain, val.Denom)
		if err != nil {
			if err == v8.Nil {
				amount = constant.ZeroDenomAmount
			}
			amount = constant.UnknownDenomAmount
		}
		_, ok := mapChainData[val.Chain+val.Denom]
		if !ok {
			mapChainData[val.Chain+val.Denom] = amount
		}

		//todo replace with val.hops
		hop := getHops(val.DenomPath)

		if val.DenomPath == "" || hop == 0 {
			continue
		}

		hopDatas, exist := mapHopsData[hop]
		if exist {
			hopDatas = append(hopDatas, val)
			mapHopsData[hop] = hopDatas
		} else {
			mapHopsData[hop] = entity.IBCDenomList{val}
		}
	}

	resp := &vo.TokenDistributionResp{
		Chain: req.BaseDenomChain,
		Denom: req.BaseDenom,
		Hops:  0,
	}
	//hop get ibc denom
	hop := 1
	resp.Children = make([]*vo.GraphData, 0, 1)
	hopDenoms, ok := mapHopsData[hop]
	if !ok {
		return resp, nil
	}
	for _, val := range hopDenoms {

		children := vo.GraphData{
			Denom:  val.Denom,
			Chain:  val.Chain,
			Hops:   hop,
			Amount: mapChainData[val.Chain+val.Denom],
		}
		children = svc.FindSource(mapChainData, mapHopsData, children)

		resp.Children = append(resp.Children, &children)
	}

	return resp, nil
}

func (svc *OverviewService) FindSource(mapChainData map[string]string, mapHopsData map[int]entity.IBCDenomList, ret vo.GraphData) vo.GraphData {
	hopDenoms, ok := mapHopsData[ret.Hops+1]
	if !ok {
		return ret
	}
	ret.Children = make([]*vo.GraphData, 0, 1)
	for _, val := range hopDenoms {
		if val.PrevDenom == ret.Denom && val.PrevChain == ret.Chain {
			children := vo.GraphData{
				Denom:  val.Denom,
				Chain:  val.Chain,
				Hops:   ret.Hops + 1,
				Amount: mapChainData[val.Chain+val.Denom],
			}
			children = svc.FindSource(mapChainData, mapHopsData, children)
			ret.Children = append(ret.Children, &children)
		}
	}

	return ret
}
