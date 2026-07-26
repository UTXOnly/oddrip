package types

import "encoding/json"

type PriceRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Step  string `json:"step"`
}

type MveSelectedLeg struct {
	EventTicker               string  `json:"event_ticker,omitempty"`
	MarketTicker              string  `json:"market_ticker,omitempty"`
	Side                      string  `json:"side,omitempty"`
	YesSettlementValueDollars *string `json:"yes_settlement_value_dollars,omitempty"`
}

type Market struct {
	Ticker                  string          `json:"ticker"`
	EventTicker             string          `json:"event_ticker"`
	MarketType              string          `json:"market_type"`
	Title                   string          `json:"title,omitempty"`
	Subtitle                string          `json:"subtitle,omitempty"`
	YesSubTitle             string          `json:"yes_sub_title"`
	NoSubTitle              string          `json:"no_sub_title"`
	CreatedTime             string          `json:"created_time"`
	UpdatedTime             string          `json:"updated_time"`
	OpenTime                string          `json:"open_time"`
	CloseTime               string          `json:"close_time"`
	ExpectedExpirationTime  *string         `json:"expected_expiration_time,omitempty"`
	ExpirationTime          string          `json:"expiration_time,omitempty"`
	LatestExpirationTime    string          `json:"latest_expiration_time"`
	SettlementTimerSeconds  int             `json:"settlement_timer_seconds"`
	Status                  string          `json:"status"`
	NotionalValueDollars    string          `json:"notional_value_dollars"`
	YesBidDollars           string          `json:"yes_bid_dollars"`
	YesAskDollars           string          `json:"yes_ask_dollars"`
	NoBidDollars            string          `json:"no_bid_dollars"`
	NoAskDollars            string          `json:"no_ask_dollars"`
	YesBidSizeFp            string          `json:"yes_bid_size_fp"`
	YesAskSizeFp            string          `json:"yes_ask_size_fp"`
	LastPriceDollars        string          `json:"last_price_dollars"`
	PreviousYesBidDollars   string          `json:"previous_yes_bid_dollars"`
	PreviousYesAskDollars   string          `json:"previous_yes_ask_dollars"`
	PreviousPriceDollars    string          `json:"previous_price_dollars"`
	VolumeFp                string          `json:"volume_fp"`
	Volume24hFp             string          `json:"volume_24h_fp"`
	LiquidityDollars        string          `json:"liquidity_dollars,omitempty"`
	OpenInterestFp          string          `json:"open_interest_fp"`
	Result                  string          `json:"result"`
	CanCloseEarly           bool            `json:"can_close_early"`
	SettlementValueDollars  *string         `json:"settlement_value_dollars,omitempty"`
	SettlementTs            *string         `json:"settlement_ts,omitempty"`
	OccurrenceDatetime      *string         `json:"occurrence_datetime,omitempty"`
	ExpirationValue         string          `json:"expiration_value"`
	FeeWaiverExpirationTime *string         `json:"fee_waiver_expiration_time,omitempty"`
	EarlyCloseCondition     string          `json:"early_close_condition,omitempty"`
	StrikeType              string          `json:"strike_type,omitempty"`
	FloorStrike             *float64        `json:"floor_strike,omitempty"`
	CapStrike               *float64        `json:"cap_strike,omitempty"`
	FunctionalStrike        *string         `json:"functional_strike,omitempty"`
	CustomStrike            json.RawMessage `json:"custom_strike,omitempty"`
	RulesPrimary            string          `json:"rules_primary"`
	RulesSecondary          string          `json:"rules_secondary"`
	PriceLevelStructure     string          `json:"price_level_structure"`
	PriceRanges             []PriceRange    `json:"price_ranges"`
	MveCollectionTicker     string          `json:"mve_collection_ticker,omitempty"`
	MveSelectedLegs         []MveSelectedLeg `json:"mve_selected_legs,omitempty"`
	PrimaryParticipantKey   *string         `json:"primary_participant_key,omitempty"`
	IsProvisional           *bool           `json:"is_provisional,omitempty"`
	ExchangeIndex           int             `json:"exchange_index,omitempty"`
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

type OrderbookCountFp struct {
	YesDollars [][2]string `json:"yes_dollars"`
	NoDollars  [][2]string `json:"no_dollars"`
}

type GetMarketOrderbookResponse struct {
	OrderbookFp OrderbookCountFp `json:"orderbook_fp"`
}

type GetMarketOrderbookOpts struct {
	Depth int
}

type GetMarketOrderbooksOpts struct {
	Tickers []string
}

type MarketOrderbookFp struct {
	Ticker      string           `json:"ticker"`
	OrderbookFp OrderbookCountFp `json:"orderbook_fp"`
}

type GetMarketOrderbooksResponse struct {
	Orderbooks []MarketOrderbookFp `json:"orderbooks"`
}

type Trade struct {
	TradeID          string `json:"trade_id"`
	Ticker           string `json:"ticker"`
	CountFp          string `json:"count_fp"`
	YesPriceDollars  string `json:"yes_price_dollars"`
	NoPriceDollars   string `json:"no_price_dollars"`
	TakerSide        string `json:"taker_side,omitempty"`
	TakerOutcomeSide string `json:"taker_outcome_side"`
	TakerBookSide    string `json:"taker_book_side"`
	CreatedTime      string `json:"created_time"`
	IsBlockTrade     bool   `json:"is_block_trade"`
}

type GetTradesResponse struct {
	Trades []Trade `json:"trades"`
	Cursor string  `json:"cursor"`
}

type GetTradesOpts struct {
	Limit        *int64
	Cursor       string
	Ticker       string
	MinTs        *int64
	MaxTs        *int64
	IsBlockTrade *bool
}
