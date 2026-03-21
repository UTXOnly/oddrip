package types

type BidAskDistributionHistorical struct {
	Open  string `json:"open"`
	Low   string `json:"low"`
	High  string `json:"high"`
	Close string `json:"close"`
}

type PriceDistributionHistorical struct {
	Open     *string `json:"open"`
	Low      *string `json:"low"`
	High     *string `json:"high"`
	Close    *string `json:"close"`
	Mean     *string `json:"mean"`
	Previous *string `json:"previous"`
}

type MarketCandlestickHistorical struct {
	EndPeriodTs int64                        `json:"end_period_ts"`
	YesBid      BidAskDistributionHistorical `json:"yes_bid"`
	YesAsk      BidAskDistributionHistorical `json:"yes_ask"`
	Price       PriceDistributionHistorical  `json:"price"`
	Volume       string `json:"volume"`
	OpenInterest string `json:"open_interest"`
}

type GetMarketCandlesticksHistoricalResponse struct {
	Ticker       string                     `json:"ticker"`
	Candlesticks []MarketCandlestickHistorical `json:"candlesticks"`
}

type GetHistoricalMarketsOpts struct {
	Limit        *int64
	Cursor       string
	Tickers      string
	EventTicker  string
	MveFilter    string
}

type GetHistoricalMarketCandlesticksOpts struct {
	StartTs        int64
	EndTs          int64
	PeriodInterval int
}

type GetHistoricalArchiveOpts struct {
	Ticker   string
	MaxTs    *int64
	Limit    *int64
	Cursor   string
}
