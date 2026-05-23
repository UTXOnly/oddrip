package types

type CreateOrderV2Request struct {
	Ticker                  string  `json:"ticker"`
	ClientOrderID           string  `json:"client_order_id"`
	Side                    string  `json:"side"`
	Count                   string  `json:"count"`
	Price                   string  `json:"price"`
	TimeInForce             string  `json:"time_in_force"`
	SelfTradePreventionType string  `json:"self_trade_prevention_type"`
	ExpirationTime          *int64  `json:"expiration_time,omitempty"`
	PostOnly                *bool   `json:"post_only,omitempty"`
	CancelOrderOnPause      *bool   `json:"cancel_order_on_pause,omitempty"`
	ReduceOnly              *bool   `json:"reduce_only,omitempty"`
	Subaccount              *int    `json:"subaccount,omitempty"`
	OrderGroupID            *string `json:"order_group_id,omitempty"`
	ExchangeIndex           *int    `json:"exchange_index,omitempty"`
}

type CreateOrderV2Response struct {
	OrderID           string  `json:"order_id"`
	ClientOrderID     string  `json:"client_order_id,omitempty"`
	FillCount         string  `json:"fill_count"`
	RemainingCount    string  `json:"remaining_count"`
	AverageFillPrice  string  `json:"average_fill_price,omitempty"`
	AverageFeePaid    string  `json:"average_fee_paid,omitempty"`
	TsMs              int64   `json:"ts_ms"`
}

type CancelOrderV2Response struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id,omitempty"`
	ReducedBy     string `json:"reduced_by"`
	TsMs          int64  `json:"ts_ms"`
}

type AmendOrderV2Request struct {
	Ticker               string  `json:"ticker"`
	Side                 string  `json:"side"`
	Price                string  `json:"price"`
	Count                string  `json:"count"`
	ClientOrderID        *string `json:"client_order_id,omitempty"`
	UpdatedClientOrderID *string `json:"updated_client_order_id,omitempty"`
	ExchangeIndex        *int    `json:"exchange_index,omitempty"`
}

type AmendOrderV2Response struct {
	OrderID          string  `json:"order_id"`
	ClientOrderID    string  `json:"client_order_id,omitempty"`
	RemainingCount   *string `json:"remaining_count,omitempty"`
	FillCount        *string `json:"fill_count,omitempty"`
	AverageFillPrice *string `json:"average_fill_price,omitempty"`
	AverageFeePaid   *string `json:"average_fee_paid,omitempty"`
	TsMs             int64   `json:"ts_ms"`
}

type DecreaseOrderV2Request struct {
	ReduceBy      *string `json:"reduce_by,omitempty"`
	ReduceTo      *string `json:"reduce_to,omitempty"`
	ExchangeIndex *int    `json:"exchange_index,omitempty"`
}

type DecreaseOrderV2Response struct {
	OrderID         string `json:"order_id"`
	ClientOrderID   string `json:"client_order_id,omitempty"`
	RemainingCount  string `json:"remaining_count"`
	TsMs            int64  `json:"ts_ms"`
}

type BatchCreateOrdersV2Request struct {
	Orders []CreateOrderV2Request `json:"orders"`
}

type BatchCreateOrdersV2IndividualResponse struct {
	OrderID           string         `json:"order_id,omitempty"`
	ClientOrderID     *string        `json:"client_order_id,omitempty"`
	FillCount         *string        `json:"fill_count,omitempty"`
	RemainingCount    *string        `json:"remaining_count,omitempty"`
	AverageFillPrice  *string        `json:"average_fill_price,omitempty"`
	AverageFeePaid    *string        `json:"average_fee_paid,omitempty"`
	TsMs              *int64         `json:"ts_ms,omitempty"`
	Error             *ErrorResponse `json:"error,omitempty"`
}

type BatchCreateOrdersV2Response struct {
	Orders []BatchCreateOrdersV2IndividualResponse `json:"orders"`
}

type BatchCancelOrdersV2RequestOrder struct {
	OrderID       string `json:"order_id"`
	Subaccount    *int   `json:"subaccount,omitempty"`
	ExchangeIndex *int   `json:"exchange_index,omitempty"`
}

type BatchCancelOrdersV2Request struct {
	Orders []BatchCancelOrdersV2RequestOrder `json:"orders"`
}

type BatchCancelOrdersV2IndividualResponse struct {
	OrderID       string         `json:"order_id,omitempty"`
	ClientOrderID *string        `json:"client_order_id,omitempty"`
	ReducedBy     *string        `json:"reduced_by,omitempty"`
	TsMs          *int64         `json:"ts_ms,omitempty"`
	Error         *ErrorResponse `json:"error,omitempty"`
}

type BatchCancelOrdersV2Response struct {
	Orders []BatchCancelOrdersV2IndividualResponse `json:"orders"`
}
