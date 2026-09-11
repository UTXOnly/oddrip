package types

type OrderGroup struct {
	ID                  string `json:"id"`
	ContractsLimitFp    string `json:"contracts_limit_fp,omitempty"`
	IsAutoCancelEnabled bool   `json:"is_auto_cancel_enabled"`
	ExchangeIndex       int    `json:"exchange_index,omitempty"`
}

type GetOrderGroupsResponse struct {
	OrderGroups []OrderGroup `json:"order_groups"`
}

// GetOrderGroupsOpts: Subaccount nil returns groups across all subaccounts.
type GetOrderGroupsOpts struct {
	Subaccount *int
}

type GetOrderGroupOpts struct {
	Subaccount *int
}

// OrderGroupOpts carries the query parameters for delete, reset, trigger, and
// limit updates. Both default to 0 (primary account, first shard) when nil.
type OrderGroupOpts struct {
	Subaccount    *int
	ExchangeIndex *int
}

// CreateOrderGroupRequest requires ContractsLimit or ContractsLimitFp; if both
// are set they must match.
type CreateOrderGroupRequest struct {
	Subaccount       *int    `json:"subaccount,omitempty"`
	ContractsLimit   *int64  `json:"contracts_limit,omitempty"`
	ContractsLimitFp *string `json:"contracts_limit_fp,omitempty"`
	ExchangeIndex    *int    `json:"exchange_index,omitempty"`
}

type CreateOrderGroupResponse struct {
	OrderGroupID  string `json:"order_group_id"`
	Subaccount    int    `json:"subaccount"`
	ExchangeIndex int    `json:"exchange_index,omitempty"`
}

type GetOrderGroupResponse struct {
	IsAutoCancelEnabled bool     `json:"is_auto_cancel_enabled"`
	ContractsLimitFp    string   `json:"contracts_limit_fp,omitempty"`
	Orders              []string `json:"orders"`
	ExchangeIndex       int      `json:"exchange_index,omitempty"`
}

type UpdateOrderGroupLimitRequest struct {
	ContractsLimit   *int64  `json:"contracts_limit,omitempty"`
	ContractsLimitFp *string `json:"contracts_limit_fp,omitempty"`
}
