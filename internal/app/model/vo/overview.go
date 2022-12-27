package vo

type MarketHeatmapResp struct {
	Items     []HeatmapItem    `json:"items"`
	TotalInfo HeatmapTotalInfo `json:"total_info"`
}

type HeatmapItem struct {
	Price               float64 `json:"price"`
	PriceGrowthRate     float64 `json:"price_growth_rate"`
	PriceTrend          string  `json:"price_trend"`
	Denom               string  `json:"denom"`
	Chain               string  `json:"chain"`
	MarketCapValue      string  `json:"market_cap_value"`
	TransferVolumeValue string  `json:"transfer_volume_value"`
}

type HeatmapTotalInfo struct {
	StablecoinsMarketCap string  `json:"stablecoins_market_cap"`
	TotalMarketCap       string  `json:"total_market_cap"`
	MarketCapGrowthRate  float64 `json:"market_cap_growth_rate"`
	MarketCapTrend       string  `json:"market_cap_trend"`
	TransferVolumeTotal  string  `json:"transfer_volume_total"`
	AtomPrice            float64 `json:"atom_price"`
	AtomDominance        float64 `json:"atom_dominance"`
}

type VolumeItem struct {
	Datetime string `json:"datetime"`
	Value    string `json:"value"`
}
