package types

// Perps (margin) WebSocket payloads, for connections from ConnectPerpsWS. The
// perps socket uses the same channel names, WSType* message types, and command
// shapes as the event-contract socket, but its msg payloads differ; decode
// them into these types. order_group_updates payloads decode into
// OrderGroupUpdatesMsg. Prices and counts are fixed-point strings; timestamps
// are Unix milliseconds.

// MarginErrorMsg is the msg of a perps error frame, whether it is a command
// rejection or scoped to a subscription (the envelope then carries its SID).
type MarginErrorMsg struct {
	Code          int      `json:"code"`
	Msg           string   `json:"msg"`
	MarketTicker  string   `json:"market_ticker,omitempty"`
	MarketTickers []string `json:"market_tickers,omitempty"`
}

// MarginOrderbookSnapshotMsg is an orderbook_snapshot payload: aggregated
// [price, count] levels per book side. Either side may be absent.
type MarginOrderbookSnapshotMsg struct {
	MarketTicker string           `json:"market_ticker"`
	Bid          []OrderbookLevel `json:"bid,omitempty"`
	Ask          []OrderbookLevel `json:"ask,omitempty"`
}

// MarginOrderbookDeltaMsg is an orderbook_delta payload. Side is BookSideBid or
// BookSideAsk. LastUpdateReason is a MarginLastUpdateReason* value, present
// only when the change comes from the authenticated user's order.
type MarginOrderbookDeltaMsg struct {
	MarketTicker     string `json:"market_ticker"`
	Price            string `json:"price"`
	Delta            string `json:"delta"`
	Side             string `json:"side"`
	LastUpdateReason string `json:"last_update_reason,omitempty"`
	ClientOrderID    string `json:"client_order_id,omitempty"`
	TsMs             *int64 `json:"ts_ms,omitempty"`
	Subaccount       *int   `json:"subaccount,omitempty"`
}

// MarginTickerMsg is a ticker payload, coalesced to at most one per market per
// second. Volume and OpenInterest are contract counts. The reference, mark,
// and funding fields are nil when the server omits them.
type MarginTickerMsg struct {
	MarketTicker                     string                   `json:"market_ticker"`
	Price                            string                   `json:"price"`
	Bid                              string                   `json:"bid"`
	Ask                              string                   `json:"ask"`
	BidSizeFp                        string                   `json:"bid_size_fp"`
	AskSizeFp                        string                   `json:"ask_size_fp"`
	LastTradeSizeFp                  string                   `json:"last_trade_size_fp"`
	Volume                           string                   `json:"volume"`
	VolumeNotionalValueDollars       string                   `json:"volume_notional_value_dollars"`
	Volume24h                        string                   `json:"volume_24h"`
	Volume24hNotionalValueDollars    string                   `json:"volume_24h_notional_value_dollars"`
	OpenInterest                     string                   `json:"open_interest"`
	OpenInterestNotionalValueDollars string                   `json:"open_interest_notional_value_dollars"`
	ReferencePrice                   *TickerPrice             `json:"reference_price,omitempty"`
	SettlementMarkPrice              *TickerPrice             `json:"settlement_mark_price,omitempty"`
	LiquidationMarkPrice             *TickerPrice             `json:"liquidation_mark_price,omitempty"`
	FundingRate                      *MarginTickerFundingRate `json:"funding_rate,omitempty"`
	TsMs                             int64                    `json:"ts_ms"`
}

// MarginTickerFundingRate is the funding snapshot on a perps ticker. Rate is a
// decimal, not a percentage.
type MarginTickerFundingRate struct {
	Rate              float64 `json:"rate"`
	NextFundingTimeMs int64   `json:"next_funding_time_ms"`
	TsMs              int64   `json:"ts_ms"`
}

// MarginTradeMsg is a public trade payload. TakerSide is BookSideBid or
// BookSideAsk.
type MarginTradeMsg struct {
	TradeID      string `json:"trade_id"`
	MarketTicker string `json:"market_ticker"`
	Price        string `json:"price"`
	Count        string `json:"count"`
	TakerSide    string `json:"taker_side"`
	TsMs         int64  `json:"ts_ms"`
}

// MarginFillMsg is a fill payload for the authenticated user. OrderSource is
// OrderSourceSystem for liquidations and exit or trailing-stop triggers,
// OrderSourceUser otherwise.
type MarginFillMsg struct {
	TradeID       string `json:"trade_id"`
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id,omitempty"`
	MarketTicker  string `json:"market_ticker"`
	IsTaker       bool   `json:"is_taker"`
	Side          string `json:"side"`
	TsMs          int64  `json:"ts_ms"`
	Price         string `json:"price"`
	Count         string `json:"count"`
	FeeCost       string `json:"fee_cost"`
	PostPosition  string `json:"post_position"`
	Subaccount    *int   `json:"subaccount,omitempty"`
	OrderSource   string `json:"order_source"`
}

// MarginUserOrderMsg is a user_order payload. CreatedTsMs is nil when the
// server sends null. LastUpdateReason is a MarginLastUpdateReason* value,
// empty when no reason applies.
type MarginUserOrderMsg struct {
	OrderID                 string `json:"order_id"`
	UserID                  string `json:"user_id"`
	ClientOrderID           string `json:"client_order_id"`
	Ticker                  string `json:"ticker"`
	Side                    string `json:"side"`
	Price                   string `json:"price"`
	FillCount               string `json:"fill_count"`
	RemainingCount          string `json:"remaining_count"`
	SelfTradePreventionType string `json:"self_trade_prevention_type,omitempty"`
	OrderGroupID            string `json:"order_group_id,omitempty"`
	ExpirationTsMs          *int64 `json:"expiration_ts_ms,omitempty"`
	CreatedTsMs             *int64 `json:"created_ts_ms,omitempty"`
	LastUpdatedTsMs         *int64 `json:"last_updated_ts_ms,omitempty"`
	LastUpdateReason        string `json:"last_update_reason,omitempty"`
	SubaccountNumber        *int   `json:"subaccount_number,omitempty"`
	OrderSource             string `json:"order_source"`
}
