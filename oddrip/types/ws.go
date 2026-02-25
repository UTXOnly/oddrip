package types

import "encoding/json"

const (
	WSChannelOrderbookDelta   = "orderbook_delta"
	WSChannelTicker           = "ticker"
	WSChannelTrade            = "trade"
	WSChannelFill             = "fill"
	WSChannelMarketPositions  = "market_positions"
	WSChannelMarketLifecycle  = "market_lifecycle_v2"
	WSChannelMultivariate    = "multivariate"
	WSChannelCommunications   = "communications"
	WSChannelOrderGroup       = "order_group_updates"
	WSChannelUserOrders       = "user_orders"
)

type SubscribeParams struct {
	Channels            []string `json:"channels"`
	MarketTicker        string   `json:"market_ticker,omitempty"`
	MarketTickers       []string `json:"market_tickers,omitempty"`
	MarketID            string   `json:"market_id,omitempty"`
	MarketIDs           []string `json:"market_ids,omitempty"`
	SendInitialSnapshot *bool    `json:"send_initial_snapshot,omitempty"`
	SkipTickerAck       *bool    `json:"skip_ticker_ack,omitempty"`
	ShardFactor         *int    `json:"shard_factor,omitempty"`
	ShardKey            *int    `json:"shard_key,omitempty"`
}

type SubscribeCommand struct {
	ID     int             `json:"id"`
	Cmd    string          `json:"cmd"`
	Params SubscribeParams `json:"params"`
}

type UnsubscribeCommand struct {
	ID     int   `json:"id"`
	Cmd    string `json:"cmd"`
	Params struct {
		Sids []int `json:"sids"`
	} `json:"params"`
}

type UpdateSubscriptionParams struct {
	SID                 *int     `json:"sid,omitempty"`
	Sids                []int    `json:"sids,omitempty"`
	MarketTicker        string   `json:"market_ticker,omitempty"`
	MarketTickers       []string `json:"market_tickers,omitempty"`
	MarketID            string   `json:"market_id,omitempty"`
	MarketIDs           []string `json:"market_ids,omitempty"`
	SendInitialSnapshot *bool    `json:"send_initial_snapshot,omitempty"`
	Action              string   `json:"action"`
}

type UpdateSubscriptionCommand struct {
	ID     int                    `json:"id"`
	Cmd    string                 `json:"cmd"`
	Params UpdateSubscriptionParams `json:"params"`
}

type ListSubscriptionsCommand struct {
	ID   int    `json:"id"`
	Cmd  string `json:"cmd"`
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
	MarketTickers []string `json:"market_tickers,omitempty"`
	MarketIDs     []string `json:"market_ids,omitempty"`
}

type OKResponse struct {
	ID   int      `json:"id,omitempty"`
	SID  int      `json:"sid,omitempty"`
	Seq  int      `json:"seq,omitempty"`
	Type string   `json:"type"`
	Msg  *OKMsg   `json:"msg,omitempty"`
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
	ID   int                      `json:"id"`
	Type string                   `json:"type"`
	Msg  []ListSubscriptionsItem `json:"msg"`
}

type WSMessage struct {
	Type string          `json:"type"`
	SID  int             `json:"sid,omitempty"`
	Seq  int             `json:"seq,omitempty"`
	Msg  json.RawMessage `json:"msg,omitempty"`
}
