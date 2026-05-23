package types

type GetBalanceResponse struct {
	Balance        int64  `json:"balance"`
	BalanceDollars string `json:"balance_dollars"`
	PortfolioValue int64  `json:"portfolio_value"`
	UpdatedTs      int64  `json:"updated_ts"`
}

type GetBalanceOpts struct {
	Subaccount *int
}

type Fill struct {
	OutcomeSide          string  `json:"outcome_side,omitempty"`
	BookSide             string  `json:"book_side,omitempty"`
	FillID           string  `json:"fill_id"`
	TradeID          string  `json:"trade_id"`
	OrderID          string  `json:"order_id"`
	ClientOrderID    string  `json:"client_order_id,omitempty"`
	Ticker           string  `json:"ticker"`
	MarketTicker     string  `json:"market_ticker"`
	Side             string  `json:"side"`
	Action           string  `json:"action"`
	Count            int     `json:"count,omitempty"`
	CountFp          string  `json:"count_fp"`
	Price            float64 `json:"price,omitempty"`
	YesPrice         int     `json:"yes_price,omitempty"`
	NoPrice          int     `json:"no_price,omitempty"`
	YesPriceDollars  string  `json:"yes_price_dollars"`
	NoPriceDollars   string  `json:"no_price_dollars"`
	YesPriceFixed    string  `json:"yes_price_fixed"`
	NoPriceFixed     string  `json:"no_price_fixed"`
	IsTaker          bool    `json:"is_taker"`
	CreatedTime      string  `json:"created_time,omitempty"`
	FeeCost          string  `json:"fee_cost"`
	Ts               *int64  `json:"ts,omitempty"`
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
	LastUpdatedTs         string `json:"last_updated_ts"`
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
	RestingOrdersCount   int    `json:"resting_orders_count,omitempty"`
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

type BucketLimit struct {
	RefillRate     int `json:"refill_rate"`
	BucketCapacity int `json:"bucket_capacity"`
}

type GetAccountApiLimitsResponse struct {
	UsageTier string      `json:"usage_tier"`
	Read      BucketLimit `json:"read"`
	Write     BucketLimit `json:"write"`
}

type EndpointTokenCost struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Cost   int    `json:"cost"`
}

type GetAccountEndpointCostsResponse struct {
	DefaultCost   int                 `json:"default_cost"`
	EndpointCosts []EndpointTokenCost `json:"endpoint_costs"`
}

type Deposit struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	AmountCents int64  `json:"amount_cents"`
	FeeCents    int64  `json:"fee_cents"`
	CreatedTs   int64  `json:"created_ts"`
}

type GetDepositsResponse struct {
	Deposits []Deposit `json:"deposits"`
	Cursor   string    `json:"cursor,omitempty"`
}

type Withdrawal struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	AmountCents int64  `json:"amount_cents"`
	FeeCents    int64  `json:"fee_cents"`
	CreatedTs   int64  `json:"created_ts"`
}

type GetWithdrawalsResponse struct {
	Withdrawals []Withdrawal `json:"withdrawals"`
	Cursor      string       `json:"cursor,omitempty"`
}

type GetDepositsOpts struct {
	Limit  *int64
	Cursor string
}

type GetWithdrawalsOpts struct {
	Limit  *int64
	Cursor string
}

type Settlement struct {
	Ticker              string `json:"ticker"`
	EventTicker         string `json:"event_ticker"`
	MarketResult        string `json:"market_result"`
	YesCountFp          string `json:"yes_count_fp"`
	YesTotalCost        int    `json:"yes_total_cost,omitempty"`
	YesTotalCostDollars string `json:"yes_total_cost_dollars"`
	NoCountFp           string `json:"no_count_fp"`
	NoTotalCost         int    `json:"no_total_cost,omitempty"`
	NoTotalCostDollars  string `json:"no_total_cost_dollars"`
	Revenue             int    `json:"revenue"`
	SettledTime         string `json:"settled_time"`
	FeeCost             string `json:"fee_cost"`
	Value               *int   `json:"value,omitempty"`
}

type GetSettlementsResponse struct {
	Settlements []Settlement `json:"settlements"`
	Cursor      string       `json:"cursor,omitempty"`
}

type GetSettlementsOpts struct {
	Limit        *int64
	Cursor       string
	Ticker       string
	EventTicker  string
	MinTs        *int64
	MaxTs        *int64
	Subaccount   *int
}
