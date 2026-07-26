package types

import "encoding/json"

const (
	WSChannelOrderbookDelta        = "orderbook_delta"
	WSChannelTicker                = "ticker"
	WSChannelTrade                 = "trade"
	WSChannelFill                  = "fill"
	WSChannelMarketPositions       = "market_positions"
	WSChannelMarketLifecycle       = "market_lifecycle_v2"
	WSChannelMultivariateLifecycle = "multivariate_market_lifecycle"
	WSChannelCommunications        = "communications"
	WSChannelOrderGroup            = "order_group_updates"
	WSChannelUserOrders            = "user_orders"
	WSChannelPythValue             = "pyth_value"
)

const (
	WSUpdateSubscriptionAddMarkets           = "add_markets"
	WSUpdateSubscriptionDeleteMarkets        = "delete_markets"
	WSUpdateSubscriptionGetSnapshot          = "get_snapshot"
	WSUpdateSubscriptionSubscribeUnderlyings = "subscribe_underlyings"
	WSUpdateSubscriptionUnsubscribeUnderlyings = "unsubscribe_underlyings"
	WSUpdateSubscriptionUnderlyingList       = "underlying_list"
)

type SubscribeParams struct {
	Channels            []string `json:"channels"`
	MarketTicker        string   `json:"market_ticker,omitempty"`
	MarketTickers       []string `json:"market_tickers,omitempty"`
	MarketID            string   `json:"market_id,omitempty"`
	MarketIDs           []string `json:"market_ids,omitempty"`
	UnderlyingTickers   []string `json:"underlying_tickers,omitempty"`
	SendInitialSnapshot *bool    `json:"send_initial_snapshot,omitempty"`
	SkipTickerAck       *bool    `json:"skip_ticker_ack,omitempty"`
	ShardFactor         *int     `json:"shard_factor,omitempty"`
	ShardKey            *int     `json:"shard_key,omitempty"`
}

type SubscribeCommand struct {
	ID     int             `json:"id"`
	Cmd    string          `json:"cmd"`
	Params SubscribeParams `json:"params"`
}

type UnsubscribeCommand struct {
	ID     int    `json:"id"`
	Cmd    string `json:"cmd"`
	Params struct {
		Sids []int `json:"sids"`
	} `json:"params"`
}

type UpdateSubscriptionParams struct {
	SID               *int     `json:"sid,omitempty"`
	Sids              []int    `json:"sids,omitempty"`
	MarketTicker      string   `json:"market_ticker,omitempty"`
	MarketTickers     []string `json:"market_tickers,omitempty"`
	MarketID          string   `json:"market_id,omitempty"`
	MarketIDs         []string `json:"market_ids,omitempty"`
	UnderlyingTickers []string `json:"underlying_tickers,omitempty"`
	SendInitialSnapshot *bool  `json:"send_initial_snapshot,omitempty"`
	Action            string   `json:"action"`
}

type UpdateSubscriptionCommand struct {
	ID     int                      `json:"id"`
	Cmd    string                   `json:"cmd"`
	Params UpdateSubscriptionParams `json:"params"`
}

type ListSubscriptionsCommand struct {
	ID  int    `json:"id"`
	Cmd string `json:"cmd"`
}

type SubscribedMsg struct {
	Channel string `json:"channel"`
	SID     int    `json:"sid"`
}

type SubscribedResponse struct {
	ID   int           `json:"id,omitempty"`
	Type string        `json:"type"`
	Msg  SubscribedMsg `json:"msg"`
}

type UnsubscribedResponse struct {
	ID   int    `json:"id,omitempty"`
	SID  int    `json:"sid"`
	Seq  int    `json:"seq"`
	Type string `json:"type"`
}

type OKMsg struct {
	MarketTickers     []string `json:"market_tickers,omitempty"`
	MarketIDs         []string `json:"market_ids,omitempty"`
	UnderlyingTickers []string `json:"underlying_tickers,omitempty"`
}

type OKResponse struct {
	ID   int    `json:"id,omitempty"`
	SID  int    `json:"sid,omitempty"`
	Seq  int    `json:"seq,omitempty"`
	Type string `json:"type"`
	Msg  *OKMsg `json:"msg,omitempty"`
}

type ErrorMsg struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	MarketID     string `json:"market_id,omitempty"`
	MarketTicker string `json:"market_ticker,omitempty"`
}

type ListSubscriptionsItem struct {
	Channel string `json:"channel"`
	SID     int    `json:"sid"`
}

type ListSubscriptionsResponse struct {
	ID   int                     `json:"id"`
	Type string                  `json:"type"`
	Msg  []ListSubscriptionsItem `json:"msg"`
}

type WSMessage struct {
	Type string          `json:"type"`
	SID  int             `json:"sid,omitempty"`
	Seq  int             `json:"seq,omitempty"`
	Msg  json.RawMessage `json:"msg,omitempty"`
}

type PythValueMsg struct {
	UnderlyingTicker string `json:"underlying_ticker"`
	ValueUSD         string `json:"value_usd"`
	SourceTsMs       int64  `json:"source_ts_ms"`
	ReceivedAt       int64  `json:"received_at"`
}

type PythUnderlyingListMsg struct {
	UnderlyingTickers []string `json:"underlying_tickers"`
}

type MarketLifecycleAdditionalMetadata struct {
	Name                 string          `json:"name,omitempty"`
	Title                string          `json:"title,omitempty"`
	YesSubTitle          string          `json:"yes_sub_title,omitempty"`
	NoSubTitle           string          `json:"no_sub_title,omitempty"`
	RulesPrimary         string          `json:"rules_primary,omitempty"`
	RulesSecondary       string          `json:"rules_secondary,omitempty"`
	CanCloseEarly        *bool           `json:"can_close_early,omitempty"`
	EventTicker          string          `json:"event_ticker,omitempty"`
	ExpectedExpirationTs *int64          `json:"expected_expiration_ts,omitempty"`
	StrikeType           string          `json:"strike_type,omitempty"`
	FloorStrike          *float64        `json:"floor_strike,omitempty"`
	CapStrike            *float64        `json:"cap_strike,omitempty"`
	CustomStrike         json.RawMessage `json:"custom_strike,omitempty"`
}

type MarketLifecycleV2Msg struct {
	EventType           string                            `json:"event_type"`
	MarketTicker        string                            `json:"market_ticker"`
	OpenTs              *int64                            `json:"open_ts,omitempty"`
	CloseTs             *int64                            `json:"close_ts,omitempty"`
	Result              string                            `json:"result,omitempty"`
	DeterminationTs     *int64                            `json:"determination_ts,omitempty"`
	SettlementValue     string                            `json:"settlement_value,omitempty"`
	SettledTs           *int64                            `json:"settled_ts,omitempty"`
	IsDeactivated       *bool                             `json:"is_deactivated,omitempty"`
	PriceLevelStructure string                            `json:"price_level_structure,omitempty"`
	PriceRanges         []PriceRange                      `json:"price_ranges,omitempty"`
	StrikeType          string                            `json:"strike_type,omitempty"`
	FloorStrike         *float64                          `json:"floor_strike,omitempty"`
	CapStrike           *float64                          `json:"cap_strike,omitempty"`
	CustomStrike        json.RawMessage                   `json:"custom_strike,omitempty"`
	YesSubTitle         string                            `json:"yes_sub_title,omitempty"`
	AdditionalMetadata  *MarketLifecycleAdditionalMetadata `json:"additional_metadata,omitempty"`
}
