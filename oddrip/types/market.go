package types

type PriceRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Step  string `json:"step"`
}

type Market struct {
	Ticker                 string       `json:"ticker"`
	EventTicker            string       `json:"event_ticker"`
	MarketType             string       `json:"market_type"`
	Title                  string       `json:"title"`
	Subtitle               string       `json:"subtitle"`
	YesSubTitle            string       `json:"yes_sub_title"`
	NoSubTitle             string       `json:"no_sub_title"`
	CreatedTime            string       `json:"created_time"`
	UpdatedTime            string       `json:"updated_time"`
	OpenTime               string       `json:"open_time"`
	CloseTime              string       `json:"close_time"`
	ExpirationTime         string       `json:"expiration_time"`
	LatestExpirationTime   string       `json:"latest_expiration_time"`
	SettlementTimerSeconds int          `json:"settlement_timer_seconds"`
	Status                 string       `json:"status"`
	ResponsePriceUnits     string       `json:"response_price_units"`
	NotionalValue          int          `json:"notional_value"`
	NotionalValueDollars   string       `json:"notional_value_dollars"`
	YesBid                 float64      `json:"yes_bid"`
	YesBidDollars          string       `json:"yes_bid_dollars"`
	YesAsk                 float64      `json:"yes_ask"`
	YesAskDollars          string       `json:"yes_ask_dollars"`
	NoBid                  float64      `json:"no_bid"`
	NoBidDollars           string       `json:"no_bid_dollars"`
	NoAsk                  float64      `json:"no_ask"`
	NoAskDollars           string       `json:"no_ask_dollars"`
	YesBidSizeFp           string       `json:"yes_bid_size_fp"`
	YesAskSizeFp           string       `json:"yes_ask_size_fp"`
	LastPrice              float64      `json:"last_price"`
	LastPriceDollars       string       `json:"last_price_dollars"`
	PreviousYesBid         int          `json:"previous_yes_bid"`
	PreviousYesBidDollars  string       `json:"previous_yes_bid_dollars"`
	PreviousYesAsk         int          `json:"previous_yes_ask"`
	PreviousYesAskDollars  string       `json:"previous_yes_ask_dollars"`
	PreviousPrice          int          `json:"previous_price"`
	PreviousPriceDollars   string       `json:"previous_price_dollars"`
	Volume                 int          `json:"volume"`
	VolumeFp               string       `json:"volume_fp"`
	Volume24h              int          `json:"volume_24h"`
	Volume24hFp            string       `json:"volume_24h_fp"`
	Liquidity              int          `json:"liquidity"`
	LiquidityDollars       string       `json:"liquidity_dollars"`
	OpenInterest           int          `json:"open_interest"`
	OpenInterestFp         string       `json:"open_interest_fp"`
	Result                 string       `json:"result"`
	CanCloseEarly          bool         `json:"can_close_early"`
	FractionalTradingEnabled bool       `json:"fractional_trading_enabled"`
	ExpirationValue        string       `json:"expiration_value"`
	RulesPrimary           string       `json:"rules_primary"`
	RulesSecondary         string       `json:"rules_secondary"`
	TickSize               int          `json:"tick_size"`
	PriceLevelStructure    string       `json:"price_level_structure"`
	PriceRanges            []PriceRange `json:"price_ranges"`
}

type GetMarketResponse struct {
	Market Market `json:"market"`
}

type GetMarketsResponse struct {
	Markets []Market `json:"markets"`
	Cursor  string   `json:"cursor"`
}

type GetMarketsOpts struct {
	Limit        *int64
	Cursor       string
	EventTicker  string
	SeriesTicker string
	MinCreatedTs *int64
	MaxCreatedTs *int64
	MinUpdatedTs *int64
	MaxCloseTs   *int64
	MinCloseTs   *int64
	MinSettledTs *int64
	MaxSettledTs *int64
	Status       string
	Tickers      string
	MveFilter    string
}

type OrderbookLevel [2]float64

type PriceLevelDollars [2]interface{}

type Orderbook struct {
	Yes        []OrderbookLevel   `json:"yes"`
	No         []OrderbookLevel   `json:"no"`
	YesDollars []PriceLevelDollars `json:"yes_dollars"`
	NoDollars  []PriceLevelDollars `json:"no_dollars"`
}

type OrderbookCountFp struct {
	YesDollars [][2]string `json:"yes_dollars"`
	NoDollars  [][2]string `json:"no_dollars"`
}

type GetMarketOrderbookResponse struct {
	Orderbook   Orderbook       `json:"orderbook"`
	OrderbookFp OrderbookCountFp `json:"orderbook_fp"`
}

type GetMarketOrderbookOpts struct {
	Depth int
}

type Trade struct {
	TradeID        string  `json:"trade_id"`
	Ticker         string  `json:"ticker"`
	Price          float64 `json:"price"`
	Count          int     `json:"count"`
	CountFp        string  `json:"count_fp"`
	YesPrice       int     `json:"yes_price"`
	NoPrice        int     `json:"no_price"`
	YesPriceDollars string `json:"yes_price_dollars"`
	NoPriceDollars  string `json:"no_price_dollars"`
	TakerSide      string  `json:"taker_side"`
	CreatedTime    string  `json:"created_time"`
}

type GetTradesResponse struct {
	Trades []Trade `json:"trades"`
	Cursor string  `json:"cursor"`
}

type GetTradesOpts struct {
	Limit  *int64
	Cursor string
	Ticker string
	MinTs  *int64
	MaxTs  *int64
}
