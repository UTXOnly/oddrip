package types

// MarginExchangeStatus is the perps exchange status. It carries only the two
// flags the perps spec defines; ExchangeStatus is the event-contract shape.
type MarginExchangeStatus struct {
	ExchangeActive bool `json:"exchange_active"`
	TradingActive  bool `json:"trading_active"`
}

type MarginEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// MarginMarketSchedule is a perps market's current trading schedule.
// NextCloseTs is nil while closed and NextOpenTs is nil while open; both are
// Unix seconds.
type MarginMarketSchedule struct {
	IsOpen      bool   `json:"is_open"`
	NextCloseTs *int64 `json:"next_close_ts,omitempty"`
	NextOpenTs  *int64 `json:"next_open_ts,omitempty"`
}

// MarginMarket is a perps market. The leverage estimate maps are keyed by
// notional position size in dollars ("1000", "10000", ...). LeverageEstimate
// and LeverageEstimates are nil when margin config or price data is
// unavailable. Schedule is nil for markets that trade 24/7.
type MarginMarket struct {
	Ticker                           string                `json:"ticker"`
	Title                            string                `json:"title"`
	ExchangeIndex                    int                   `json:"exchange_index"`
	ContractSize                     string                `json:"contract_size"`
	UnderlyingMultiplier             string                `json:"underlying_multiplier"`
	TickSize                         string                `json:"tick_size"`
	Status                           string                `json:"status"`
	FractionalTradingEnabled         bool                  `json:"fractional_trading_enabled"`
	LeverageEstimate                 *float64              `json:"leverage_estimate,omitempty"`
	LeverageEstimates                map[string]float64    `json:"leverage_estimates,omitempty"`
	LongLeverageEstimates            map[string]float64    `json:"long_leverage_estimates,omitempty"`
	ShortLeverageEstimates           map[string]float64    `json:"short_leverage_estimates,omitempty"`
	Price                            string                `json:"price,omitempty"`
	Volume                           string                `json:"volume,omitempty"`
	VolumeNotionalValueDollars       string                `json:"volume_notional_value_dollars,omitempty"`
	OpenInterest                     string                `json:"open_interest,omitempty"`
	OpenInterestNotionalValueDollars string                `json:"open_interest_notional_value_dollars,omitempty"`
	Volume24h                        string                `json:"volume_24h,omitempty"`
	Volume24hNotionalValueDollars    string                `json:"volume_24h_notional_value_dollars,omitempty"`
	Bid                              string                `json:"bid,omitempty"`
	Ask                              string                `json:"ask,omitempty"`
	SettlementMarkPrice              *TickerPrice          `json:"settlement_mark_price,omitempty"`
	LiquidationMarkPrice             *TickerPrice          `json:"liquidation_mark_price,omitempty"`
	ReferencePrice                   *TickerPrice          `json:"reference_price,omitempty"`
	AssetClass                       string                `json:"asset_class,omitempty"`
	ProductMetadata                  map[string]any        `json:"product_metadata,omitempty"`
	Schedule                         *MarginMarketSchedule `json:"schedule,omitempty"`
}

type MarginMarketResponse struct {
	Market MarginMarket `json:"market"`
}

type GetMarginMarketsResponse struct {
	Markets []MarginMarket `json:"markets"`
}

// GetMarginMarketsOpts: Status is one of the MarginMarketStatus* values.
type GetMarginMarketsOpts struct {
	Status string
}

// MarginOrderbookCount holds bids best-first (descending) and asks best-first
// (ascending).
type MarginOrderbookCount struct {
	Bids []OrderbookLevel `json:"bids"`
	Asks []OrderbookLevel `json:"asks"`
}

type MarginOrderbookResponse struct {
	Orderbook MarginOrderbookCount `json:"orderbook"`
}

// GetMarginMarketOrderbookOpts: Depth 0 means all levels. AggregationTickSize
// is a dollar string (e.g. "0.10") that buckets price levels.
type GetMarginMarketOrderbookOpts struct {
	Depth               *int
	AggregationTickSize string
}

type MarginMarketCandlestick struct {
	EndPeriodTs                      int64                        `json:"end_period_ts"`
	Bid                              BidAskDistributionHistorical `json:"bid"`
	Ask                              BidAskDistributionHistorical `json:"ask"`
	Price                            PriceDistributionHistorical  `json:"price"`
	Volume                           string                       `json:"volume"`
	VolumeNotionalValueDollars       string                       `json:"volume_notional_value_dollars"`
	OpenInterest                     string                       `json:"open_interest"`
	OpenInterestNotionalValueDollars string                       `json:"open_interest_notional_value_dollars"`
}

type GetMarginMarketCandlesticksResponse struct {
	Ticker       string                    `json:"ticker"`
	Candlesticks []MarginMarketCandlestick `json:"candlesticks"`
}

// GetMarginMarketCandlesticksOpts: StartTs, EndTs, and PeriodInterval are
// required; PeriodInterval is 1, 60, or 1440 (PeriodInterval* constants).
type GetMarginMarketCandlesticksOpts struct {
	StartTs                  int64
	EndTs                    int64
	PeriodInterval           int
	IncludeLatestBeforeStart *bool
}

// MarginTrade is a public perps trade. TakerSide is BookSideBid or
// BookSideAsk.
type MarginTrade struct {
	TradeID     string `json:"trade_id"`
	Ticker      string `json:"ticker"`
	Count       string `json:"count"`
	Price       string `json:"price"`
	CreatedTime string `json:"created_time"`
	TakerSide   string `json:"taker_side"`
}

type GetMarginTradesResponse struct {
	Trades []MarginTrade `json:"trades"`
	Cursor string        `json:"cursor"`
}

// GetMarginTradesOpts: Ticker is required. Limit is 1-1000 (server default
// 100).
type GetMarginTradesOpts struct {
	Ticker string
	Limit  *int64
	Cursor string
	MinTs  *int64
	MaxTs  *int64
}
