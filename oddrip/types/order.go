package types

type Order struct {
	OrderID              string  `json:"order_id"`
	UserID               string  `json:"user_id"`
	ClientOrderID        string  `json:"client_order_id"`
	Ticker               string  `json:"ticker"`
	Side                 string  `json:"side"`
	Action               string  `json:"action"`
	Type                 string  `json:"type"`
	Status               string  `json:"status"`
	YesPrice             int     `json:"yes_price"`
	NoPrice              int     `json:"no_price"`
	YesPriceDollars      string  `json:"yes_price_dollars"`
	NoPriceDollars       string  `json:"no_price_dollars"`
	FillCount            int     `json:"fill_count"`
	FillCountFp          string  `json:"fill_count_fp"`
	RemainingCount       int     `json:"remaining_count"`
	RemainingCountFp     string  `json:"remaining_count_fp"`
	InitialCount         int     `json:"initial_count"`
	InitialCountFp       string  `json:"initial_count_fp"`
	TakerFees            int     `json:"taker_fees"`
	MakerFees            int     `json:"maker_fees"`
	TakerFillCost        int     `json:"taker_fill_cost"`
	MakerFillCost        int     `json:"maker_fill_cost"`
	TakerFillCostDollars  string `json:"taker_fill_cost_dollars"`
	MakerFillCostDollars  string `json:"maker_fill_cost_dollars"`
	QueuePosition        int     `json:"queue_position"`
	CreatedTime          *string `json:"created_time,omitempty"`
	LastUpdateTime       *string `json:"last_update_time,omitempty"`
	OrderGroupID         *string `json:"order_group_id,omitempty"`
	SubaccountNumber     *int    `json:"subaccount_number,omitempty"`
}

type CreateOrderRequest struct {
	Ticker                string  `json:"ticker"`
	Side                  string  `json:"side"`
	Action                string  `json:"action"`
	ClientOrderID         *string `json:"client_order_id,omitempty"`
	Count                 *int    `json:"count,omitempty"`
	CountFp               *string `json:"count_fp,omitempty"`
	YesPrice              *int    `json:"yes_price,omitempty"`
	NoPrice               *int    `json:"no_price,omitempty"`
	YesPriceDollars       *string `json:"yes_price_dollars,omitempty"`
	NoPriceDollars        *string `json:"no_price_dollars,omitempty"`
	ExpirationTs          *int64  `json:"expiration_ts,omitempty"`
	TimeInForce           *string `json:"time_in_force,omitempty"`
	BuyMaxCost            *int    `json:"buy_max_cost,omitempty"`
	PostOnly              *bool   `json:"post_only,omitempty"`
	ReduceOnly            *bool   `json:"reduce_only,omitempty"`
	SelfTradePreventionType *string `json:"self_trade_prevention_type,omitempty"`
	OrderGroupID          *string `json:"order_group_id,omitempty"`
	CancelOrderOnPause    *bool   `json:"cancel_order_on_pause,omitempty"`
	Subaccount            *int    `json:"subaccount,omitempty"`
}

type CreateOrderResponse struct {
	Order Order `json:"order"`
}

type GetOrderResponse struct {
	Order Order `json:"order"`
}

type GetOrdersResponse struct {
	Orders []Order `json:"orders"`
	Cursor string  `json:"cursor"`
}

type GetOrdersOpts struct {
	Ticker      string
	EventTicker string
	MinTs       *int64
	MaxTs       *int64
	Status      string
	Limit       *int64
	Cursor      string
	Subaccount  *int
}

type CancelOrderResponse struct {
	Order      Order  `json:"order"`
	ReducedBy  int    `json:"reduced_by"`
	ReducedByFp string `json:"reduced_by_fp"`
}

type AmendOrderRequest struct {
	Subaccount          *int    `json:"subaccount,omitempty"`
	Ticker              string  `json:"ticker"`
	Side                string  `json:"side"`
	Action              string  `json:"action"`
	ClientOrderID       *string `json:"client_order_id,omitempty"`
	UpdatedClientOrderID *string `json:"updated_client_order_id,omitempty"`
	YesPrice            *int    `json:"yes_price,omitempty"`
	NoPrice             *int    `json:"no_price,omitempty"`
	YesPriceDollars      *string `json:"yes_price_dollars,omitempty"`
	NoPriceDollars      *string `json:"no_price_dollars,omitempty"`
	Count               *int    `json:"count,omitempty"`
	CountFp             *string `json:"count_fp,omitempty"`
}

type AmendOrderResponse struct {
	OldOrder Order `json:"old_order"`
	Order    Order `json:"order"`
}

type DecreaseOrderRequest struct {
	Subaccount *int    `json:"subaccount,omitempty"`
	ReduceBy   *int    `json:"reduce_by,omitempty"`
	ReduceByFp *string `json:"reduce_by_fp,omitempty"`
	ReduceTo   *int    `json:"reduce_to,omitempty"`
	ReduceToFp *string `json:"reduce_to_fp,omitempty"`
}

type DecreaseOrderResponse struct {
	Order Order `json:"order"`
}

type OrderQueuePosition struct {
	OrderID        string `json:"order_id"`
	MarketTicker   string `json:"market_ticker"`
	QueuePosition  int    `json:"queue_position"`
	QueuePositionFp string `json:"queue_position_fp,omitempty"`
}

type GetOrderQueuePositionResponse struct {
	QueuePosition   int    `json:"queue_position"`
	QueuePositionFp string `json:"queue_position_fp,omitempty"`
}

type GetOrderQueuePositionsResponse struct {
	QueuePositions []OrderQueuePosition `json:"queue_positions"`
}

type GetOrderQueuePositionsOpts struct {
	MarketTickers string
	EventTicker   string
	Subaccount    *int
}

type BatchCreateOrdersRequest struct {
	Orders []CreateOrderRequest `json:"orders"`
}

type BatchCreateOrdersIndividualResponse struct {
	ClientOrderID *string         `json:"client_order_id,omitempty"`
	Order         *Order          `json:"order,omitempty"`
	Error         *ErrorResponse  `json:"error,omitempty"`
}

type BatchCreateOrdersResponse struct {
	Orders []BatchCreateOrdersIndividualResponse `json:"orders"`
}

type BatchCancelOrdersRequestOrder struct {
	OrderID    string `json:"order_id"`
	Subaccount *int   `json:"subaccount,omitempty"`
}

type BatchCancelOrdersRequest struct {
	Orders []BatchCancelOrdersRequestOrder `json:"orders,omitempty"`
}

type BatchCancelOrdersIndividualResponse struct {
	OrderID     string         `json:"order_id"`
	Order       *Order         `json:"order,omitempty"`
	ReducedBy   int            `json:"reduced_by"`
	ReducedByFp string         `json:"reduced_by_fp"`
	Error       *ErrorResponse `json:"error,omitempty"`
}

type BatchCancelOrdersResponse struct {
	Orders []BatchCancelOrdersIndividualResponse `json:"orders"`
}
