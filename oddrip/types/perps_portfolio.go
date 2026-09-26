package types

// GetMarginBalanceOpts: ComputeAvailableBalance fills in
// MarginSubaccountBalance.AvailableBalance and RestingOrdersMargin, which are
// otherwise 0. Kalshi charges it at a higher rate-limit cost because it scans
// all resting orders.
type GetMarginBalanceOpts struct {
	ComputeAvailableBalance *bool
}

// MarginSubaccountBalance is one subaccount's perps balance breakdown. All
// amounts are fixed-point dollar strings.
type MarginSubaccountBalance struct {
	Subaccount          int    `json:"subaccount"`
	PositionValue       string `json:"position_value"`
	AccountEquity       string `json:"account_equity"`
	MaintenanceMargin   string `json:"maintenance_margin"`
	InitialMargin       string `json:"initial_margin"`
	RestingOrdersMargin string `json:"resting_orders_margin"`
	AvailableBalance    string `json:"available_balance"`
}

type GetMarginBalanceResponse struct {
	SubaccountBalances []MarginSubaccountBalance `json:"subaccount_balances"`
	SettledFunds       string                    `json:"settled_funds"`
}

type GetMarginFillsOpts struct {
	Subaccount *int
	Limit      *int64
	Cursor     string
	MinTs      *int64
	MaxTs      *int64
}

// MarginFill is one perps fill. Side is BookSideBid or BookSideAsk; Count is a
// fixed-point count string; Price, EntryPrice, Fees, and RealizedPnl are
// fixed-point dollar strings. OrderSource is an OrderSource* value.
type MarginFill struct {
	FillID      string `json:"fill_id"`
	OrderID     string `json:"order_id"`
	IsTaker     bool   `json:"is_taker"`
	Side        string `json:"side"`
	Count       string `json:"count"`
	CreatedTime string `json:"created_time"`
	Ticker      string `json:"ticker"`
	Price       string `json:"price"`
	EntryPrice  string `json:"entry_price"`
	Fees        string `json:"fees"`
	RealizedPnl string `json:"realized_pnl"`
	OrderSource string `json:"order_source,omitempty"`
}

type GetMarginFillsResponse struct {
	Fills  []MarginFill `json:"fills"`
	Cursor string       `json:"cursor"`
}

type GetMarginPositionsOpts struct {
	Subaccount *int
	Ticker     string
}

// MarginPosition is an open perps position. Position is a signed fixed-point
// count (negative = short). MarginUsed and ROE are nil when margin cannot be
// attributed to this market alone (IsPortfolio is true) or, for ROE, when
// margin used is zero.
type MarginPosition struct {
	Subaccount    int      `json:"subaccount"`
	MarketTicker  string   `json:"market_ticker"`
	Position      string   `json:"position"`
	EntryPrice    string   `json:"entry_price"`
	UnrealizedPnl string   `json:"unrealized_pnl"`
	MarginUsed    *string  `json:"margin_used,omitempty"`
	Fees          string   `json:"fees"`
	ROE           *float64 `json:"roe,omitempty"`
	IsPortfolio   bool     `json:"is_portfolio"`
}

type GetMarginPositionsResponse struct {
	Positions []MarginPosition `json:"positions"`
}

// ApplyMarginSubaccountTransferRequest moves AmountCents between perps
// subaccounts; 0 is the primary account, 1-63 are numbered subaccounts.
// Kalshi deduplicates on ClientTransferID (a UUID).
type ApplyMarginSubaccountTransferRequest struct {
	ClientTransferID string `json:"client_transfer_id"`
	FromSubaccount   int    `json:"from_subaccount"`
	ToSubaccount     int    `json:"to_subaccount"`
	AmountCents      int64  `json:"amount_cents"`
}
