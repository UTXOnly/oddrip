package types

type Order struct {
	OrderID                   string  `json:"order_id"`
	UserID                    string  `json:"user_id"`
	ClientOrderID             string  `json:"client_order_id"`
	Ticker                    string  `json:"ticker"`
	Side                      string  `json:"side,omitempty"`
	Action                    string  `json:"action,omitempty"`
	OutcomeSide               string  `json:"outcome_side"`
	BookSide                  string  `json:"book_side"`
	Type                      string  `json:"type"`
	Status                    string  `json:"status"`
	YesPriceDollars           string  `json:"yes_price_dollars"`
	NoPriceDollars            string  `json:"no_price_dollars"`
	FillCountFp               string  `json:"fill_count_fp"`
	RemainingCountFp          string  `json:"remaining_count_fp"`
	InitialCountFp            string  `json:"initial_count_fp"`
	TakerFeesDollars          string  `json:"taker_fees_dollars"`
	MakerFeesDollars          string  `json:"maker_fees_dollars"`
	TakerFillCostDollars      string  `json:"taker_fill_cost_dollars"`
	MakerFillCostDollars      string  `json:"maker_fill_cost_dollars"`
	ExpirationTime            *string `json:"expiration_time,omitempty"`
	CreatedTime               *string `json:"created_time,omitempty"`
	LastUpdateTime            *string `json:"last_update_time,omitempty"`
	SelfTradePreventionType   *string `json:"self_trade_prevention_type,omitempty"`
	OrderGroupID              *string `json:"order_group_id,omitempty"`
	CancelOrderOnPause        *bool   `json:"cancel_order_on_pause,omitempty"`
	SubaccountNumber          *int    `json:"subaccount_number,omitempty"`
	ExchangeIndex             int     `json:"exchange_index,omitempty"`
}

type CreateOrderRequest struct {
	Ticker                  string  `json:"ticker"`
	Side                    string  `json:"side"`
	Action                  string  `json:"action"`
	ClientOrderID           *string `json:"client_order_id,omitempty"`
	Count                   *int    `json:"count,omitempty"`
	CountFp                 *string `json:"count_fp,omitempty"`
	YesPrice                *int    `json:"yes_price,omitempty"`
	NoPrice                 *int    `json:"no_price,omitempty"`
	YesPriceDollars         *string `json:"yes_price_dollars,omitempty"`
	NoPriceDollars          *string `json:"no_price_dollars,omitempty"`
	ExpirationTs            *int64  `json:"expiration_ts,omitempty"`
	TimeInForce             *string `json:"time_in_force,omitempty"`
	BuyMaxCost              *int    `json:"buy_max_cost,omitempty"`
	PostOnly                *bool   `json:"post_only,omitempty"`
	ReduceOnly              *bool   `json:"reduce_only,omitempty"`
	SelfTradePreventionType *string `json:"self_trade_prevention_type,omitempty"`
	OrderGroupID            *string `json:"order_group_id,omitempty"`
	CancelOrderOnPause      *bool   `json:"cancel_order_on_pause,omitempty"`
	Subaccount              *int    `json:"subaccount,omitempty"`
	ExchangeIndex           *int    `json:"exchange_index,omitempty"`
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
	Order       Order  `json:"order"`
	ReducedBy   int    `json:"reduced_by"`
	ReducedByFp string `json:"reduced_by_fp"`
}

type AmendOrderRequest struct {
	Subaccount           *int    `json:"subaccount,omitempty"`
	Ticker               string  `json:"ticker"`
	Side                 string  `json:"side"`
	Action               string  `json:"action"`
	ClientOrderID        *string `json:"client_order_id,omitempty"`
	UpdatedClientOrderID *string `json:"updated_client_order_id,omitempty"`
	YesPrice             *int    `json:"yes_price,omitempty"`
	NoPrice              *int    `json:"no_price,omitempty"`
	YesPriceDollars      *string `json:"yes_price_dollars,omitempty"`
	NoPriceDollars       *string `json:"no_price_dollars,omitempty"`
	Count                *int    `json:"count,omitempty"`
	CountFp              *string `json:"count_fp,omitempty"`
	ExchangeIndex        *int    `json:"exchange_index,omitempty"`
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
	OrderID         string `json:"order_id"`
	MarketTicker    string `json:"market_ticker"`
	QueuePosition   int    `json:"queue_position,omitempty"`
	QueuePositionFp string `json:"queue_position_fp"`
}

type GetOrderQueuePositionResponse struct {
	QueuePosition   int    `json:"queue_position,omitempty"`
	QueuePositionFp string `json:"queue_position_fp"`
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
	ClientOrderID *string        `json:"client_order_id,omitempty"`
	Order         *Order         `json:"order,omitempty"`
	Error         *ErrorResponse `json:"error,omitempty"`
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
