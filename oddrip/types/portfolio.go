package types

type GetBalanceResponse struct {
	Balance        int64 `json:"balance"`
	PortfolioValue int64 `json:"portfolio_value"`
	UpdatedTs      int64 `json:"updated_ts"`
}

type GetBalanceOpts struct {
	Subaccount *int
}

type Fill struct {
	FillID           string  `json:"fill_id"`
	TradeID          string  `json:"trade_id"`
	OrderID          string  `json:"order_id"`
	ClientOrderID    string  `json:"client_order_id,omitempty"`
	Ticker           string  `json:"ticker"`
	MarketTicker     string  `json:"market_ticker"`
	Side             string  `json:"side"`
	Action           string  `json:"action"`
	Count            int     `json:"count"`
	CountFp          string  `json:"count_fp"`
	Price            float64 `json:"price"`
	YesPrice         int     `json:"yes_price"`
	NoPrice          int     `json:"no_price"`
	YesPriceFixed    string  `json:"yes_price_fixed"`
	NoPriceFixed     string  `json:"no_price_fixed"`
	IsTaker          bool    `json:"is_taker"`
	CreatedTime      string  `json:"created_time,omitempty"`
	FeeCost          string  `json:"fee_cost"`
	SubaccountNumber *int    `json:"subaccount_number,omitempty"`
}

type GetFillsResponse struct {
	Fills  []Fill `json:"fills"`
	Cursor string `json:"cursor"`
}

type GetFillsOpts struct {
	Ticker     string
	OrderID    string
	MinTs      *int64
	MaxTs      *int64
	Limit      *int64
	Cursor     string
	Subaccount *int
}

type MarketPosition struct {
	Ticker                string `json:"ticker"`
	TotalTraded           int    `json:"total_traded"`
	TotalTradedDollars    string `json:"total_traded_dollars"`
	Position              int    `json:"position"`
	PositionFp            string `json:"position_fp"`
	MarketExposure        int    `json:"market_exposure"`
	MarketExposureDollars string `json:"market_exposure_dollars"`
	RealizedPnl           int    `json:"realized_pnl"`
	RealizedPnlDollars    string `json:"realized_pnl_dollars"`
	RestingOrdersCount    int    `json:"resting_orders_count"`
	FeesPaid              int    `json:"fees_paid"`
	FeesPaidDollars       string `json:"fees_paid_dollars"`
	LastUpdatedTs         string `json:"last_updated_ts,omitempty"`
}

type EventPosition struct {
	EventTicker          string `json:"event_ticker"`
	TotalCost            int    `json:"total_cost"`
	TotalCostDollars     string `json:"total_cost_dollars"`
	TotalCostShares      int64  `json:"total_cost_shares"`
	TotalCostSharesFp    string `json:"total_cost_shares_fp"`
	EventExposure        int    `json:"event_exposure"`
	EventExposureDollars string `json:"event_exposure_dollars"`
	RealizedPnl          int    `json:"realized_pnl"`
	RealizedPnlDollars   string `json:"realized_pnl_dollars"`
	RestingOrdersCount   int    `json:"resting_orders_count"`
	FeesPaid             int    `json:"fees_paid"`
	FeesPaidDollars      string `json:"fees_paid_dollars"`
}

type GetPositionsResponse struct {
	Cursor          string           `json:"cursor"`
	MarketPositions []MarketPosition `json:"market_positions"`
	EventPositions  []EventPosition  `json:"event_positions"`
}

type GetPositionsOpts struct {
	Cursor      string
	Limit       *int
	CountFilter string
	Ticker      string
	EventTicker string
	Subaccount  *int
}

type GetAccountApiLimitsResponse struct {
	UsageTier  string `json:"usage_tier"`
	ReadLimit  int    `json:"read_limit"`
	WriteLimit int    `json:"write_limit"`
}
