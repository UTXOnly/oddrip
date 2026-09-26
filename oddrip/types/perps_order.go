package types

// Perps (margin) orders. Perps order groups reuse OrderGroup,
// GetOrderGroupsResponse, GetOrderGroupResponse, CreateOrderGroupRequest,
// CreateOrderGroupResponse, and UpdateOrderGroupLimitRequest, whose shapes
// match the perps spec field for field.

// MarginOrder is a perps order. Side is BookSideBid or BookSideAsk. The
// timestamps are RFC 3339 (see ParseTime) and nil when the server sends null.
// LastUpdateReason is one of the MarginLastUpdateReason values, or empty.
// OrderSource and OrderReason mark orders the exchange placed for the user.
type MarginOrder struct {
	OrderID                 string  `json:"order_id"`
	UserID                  string  `json:"user_id"`
	ClientOrderID           string  `json:"client_order_id"`
	Ticker                  string  `json:"ticker"`
	Side                    string  `json:"side"`
	LastUpdateReason        string  `json:"last_update_reason"`
	Price                   string  `json:"price"`
	FillCount               string  `json:"fill_count"`
	RemainingCount          string  `json:"remaining_count"`
	ExpirationTime          *string `json:"expiration_time,omitempty"`
	CreatedTime             *string `json:"created_time,omitempty"`
	LastUpdateTime          *string `json:"last_update_time,omitempty"`
	SelfTradePreventionType *string `json:"self_trade_prevention_type,omitempty"`
	CancelOrderOnPause      *bool   `json:"cancel_order_on_pause,omitempty"`
	OrderGroupID            string  `json:"order_group_id,omitempty"`
	OrderSource             string  `json:"order_source,omitempty"`
	OrderReason             string  `json:"order_reason,omitempty"`
}

type GetMarginOrderResponse struct {
	Order MarginOrder `json:"order"`
}

type GetMarginOrdersResponse struct {
	Orders []MarginOrder `json:"orders"`
	Cursor string        `json:"cursor"`
}

// GetMarginOrdersOpts carries the query parameters for listing perps orders.
// Limit accepts 1-10000 and the server defaults to 10000, a wider page than
// other list endpoints. Subaccount nil returns orders across all subaccounts.
type GetMarginOrdersOpts struct {
	Ticker     string
	MinTs      *int64
	MaxTs      *int64
	Status     string
	Limit      *int64
	Cursor     string
	Subaccount *int
}

// MarginOrderOpts carries the query parameter for cancelling, decreasing, or
// amending a perps order. Subaccount nil means the primary account (0).
type MarginOrderOpts struct {
	Subaccount *int
}

// CreateMarginOrderRequest places a perps order. The spec requires
// ClientOrderID. Side is BookSideBid or BookSideAsk; TimeInForce and
// SelfTradePreventionType take the TimeInForce* and SelfTrade* constants.
// ReduceOnly orders are rejected unless TimeInForce is IOC or FOK.
type CreateMarginOrderRequest struct {
	Ticker                  string `json:"ticker"`
	ClientOrderID           string `json:"client_order_id"`
	Side                    string `json:"side"`
	Count                   string `json:"count"`
	Price                   string `json:"price"`
	TimeInForce             string `json:"time_in_force"`
	SelfTradePreventionType string `json:"self_trade_prevention_type"`
	ExpirationTime          int64  `json:"expiration_time,omitempty"`
	PostOnly                *bool  `json:"post_only,omitempty"`
	CancelOrderOnPause      *bool  `json:"cancel_order_on_pause,omitempty"`
	ReduceOnly              *bool  `json:"reduce_only,omitempty"`
	Subaccount              *int   `json:"subaccount,omitempty"`
	OrderGroupID            string `json:"order_group_id,omitempty"`
}

// CreateMarginOrderResponse: AverageFillPrice and AverageFeePaid are set only
// when FillCount is above zero.
type CreateMarginOrderResponse struct {
	OrderID          string `json:"order_id"`
	ClientOrderID    string `json:"client_order_id,omitempty"`
	FillCount        string `json:"fill_count"`
	RemainingCount   string `json:"remaining_count"`
	AverageFillPrice string `json:"average_fill_price,omitempty"`
	AverageFeePaid   string `json:"average_fee_paid,omitempty"`
}

type CancelMarginOrderResponse struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id,omitempty"`
	ReducedBy     string `json:"reduced_by"`
}

// DecreaseMarginOrderRequest requires exactly one of ReduceBy or ReduceTo.
type DecreaseMarginOrderRequest struct {
	ReduceBy *string `json:"reduce_by,omitempty"`
	ReduceTo *string `json:"reduce_to,omitempty"`
}

type DecreaseMarginOrderResponse struct {
	OrderID        string `json:"order_id"`
	ClientOrderID  string `json:"client_order_id,omitempty"`
	RemainingCount string `json:"remaining_count"`
}

// AmendMarginOrderRequest sets the order's new price and size. ClientOrderID
// is the order's current client ID; UpdatedClientOrderID replaces it.
type AmendMarginOrderRequest struct {
	Ticker               string `json:"ticker"`
	Side                 string `json:"side"`
	Price                string `json:"price"`
	Count                string `json:"count"`
	ClientOrderID        string `json:"client_order_id,omitempty"`
	UpdatedClientOrderID string `json:"updated_client_order_id,omitempty"`
}

// AmendMarginOrderResponse: RemainingCount and FillCount are set when the
// amend filled or changed the resting size; the averages only when it filled.
type AmendMarginOrderResponse struct {
	OrderID          string  `json:"order_id"`
	ClientOrderID    string  `json:"client_order_id,omitempty"`
	RemainingCount   *string `json:"remaining_count,omitempty"`
	FillCount        *string `json:"fill_count,omitempty"`
	AverageFillPrice *string `json:"average_fill_price,omitempty"`
	AverageFeePaid   *string `json:"average_fee_paid,omitempty"`
}

// MarginOrderGroupOpts carries the query parameter for deleting, resetting,
// triggering, or updating the limit of a perps order group. Subaccount nil
// means the primary account (0). The perps endpoints take no exchange_index,
// so OrderGroupOpts does not apply.
type MarginOrderGroupOpts struct {
	Subaccount *int
}
