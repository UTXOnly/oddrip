package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func (s *PerpsOrdersService) List(ctx context.Context, opts *types.GetMarginOrdersOpts) (*types.GetMarginOrdersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQuery(v, "status", opts.Status)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetMarginOrdersResponse
	if err := s.client.get(ctx, joinPath("margin", "orders"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create places a perps order. As with OrdersService.CreateV2, a non-empty
// ClientOrderID is the deduplication key: when it is set the call is retried
// on 5xx and transport errors like any idempotent request, otherwise only on
// 429 so a dropped connection cannot place the order twice.
func (s *PerpsOrdersService) Create(ctx context.Context, req *types.CreateMarginOrderRequest) (*types.CreateMarginOrderResponse, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	post := s.client.post
	if req.ClientOrderID != "" {
		post = s.client.postIdempotent
	}
	var out types.CreateMarginOrderResponse
	if err := post(ctx, joinPath("margin", "orders"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelAll cancels every resting perps order for the authenticated member.
// With opts nil or Subaccount unset, orders from any subaccount are eligible.
// Orders placed during the minute after the request may also be cancelled.
func (s *PerpsOrdersService) CancelAll(ctx context.Context, opts *types.CancelAllOrdersOpts) error {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	return s.client.delete(ctx, joinPath("margin", "orders"), v, nil, nil)
}

func (s *PerpsOrdersService) Get(ctx context.Context, orderID string) (*types.GetMarginOrderResponse, error) {
	var out types.GetMarginOrderResponse
	if err := s.client.get(ctx, joinPath("margin", "orders", orderID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels the order's remaining resting contracts.
func (s *PerpsOrdersService) Cancel(ctx context.Context, orderID string, opts *types.MarginOrderOpts) (*types.CancelMarginOrderResponse, error) {
	var out types.CancelMarginOrderResponse
	if err := s.client.delete(ctx, joinPath("margin", "orders", orderID), perpsOrderQuery(opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Decrease reduces the order's size by or to a contract count. It is retried
// on 429 only.
func (s *PerpsOrdersService) Decrease(ctx context.Context, orderID string, req *types.DecreaseMarginOrderRequest, opts *types.MarginOrderOpts) (*types.DecreaseMarginOrderResponse, error) {
	if req == nil || (req.ReduceBy == nil) == (req.ReduceTo == nil) {
		return nil, errors.New("exactly one of reduce_by or reduce_to required")
	}
	var out types.DecreaseMarginOrderResponse
	if err := s.client.postQuery(ctx, joinPath("margin", "orders", orderID, "decrease"), perpsOrderQuery(opts), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Amend changes the order's price and/or size. Only a size decrease keeps
// queue position. It is retried on 429 only.
func (s *PerpsOrdersService) Amend(ctx context.Context, orderID string, req *types.AmendMarginOrderRequest, opts *types.MarginOrderOpts) (*types.AmendMarginOrderResponse, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	var out types.AmendMarginOrderResponse
	if err := s.client.postQuery(ctx, joinPath("margin", "orders", orderID, "amend"), perpsOrderQuery(opts), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func perpsOrderQuery(opts *types.MarginOrderOpts) url.Values {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	return v
}

func (s *PerpsOrderGroupsService) List(ctx context.Context, opts *types.GetOrderGroupsOpts) (*types.GetOrderGroupsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetOrderGroupsResponse
	if err := s.client.get(ctx, joinPath("margin", "order_groups"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create makes a perps order group. It is retried on 429 only.
func (s *PerpsOrderGroupsService) Create(ctx context.Context, req *types.CreateOrderGroupRequest) (*types.CreateOrderGroupResponse, error) {
	if req == nil || (req.ContractsLimit == nil && req.ContractsLimitFp == nil) {
		return nil, errors.New("contracts_limit or contracts_limit_fp required")
	}
	var out types.CreateOrderGroupResponse
	if err := s.client.post(ctx, joinPath("margin", "order_groups", "create"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsOrderGroupsService) Get(ctx context.Context, orderGroupID string, opts *types.GetOrderGroupOpts) (*types.GetOrderGroupResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetOrderGroupResponse
	if err := s.client.get(ctx, joinPath("margin", "order_groups", orderGroupID), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes the order group and cancels every order in it.
func (s *PerpsOrderGroupsService) Delete(ctx context.Context, orderGroupID string, opts *types.MarginOrderGroupOpts) error {
	return s.client.delete(ctx, joinPath("margin", "order_groups", orderGroupID), perpsOrderGroupQuery(opts), nil, nil)
}

// Reset zeroes the group's matched-contracts counter so new orders can be placed.
func (s *PerpsOrderGroupsService) Reset(ctx context.Context, orderGroupID string, opts *types.MarginOrderGroupOpts) error {
	return s.client.put(ctx, joinPath("margin", "order_groups", orderGroupID, "reset"), perpsOrderGroupQuery(opts), nil, nil)
}

// Trigger cancels every order in the group and blocks new orders until Reset.
func (s *PerpsOrderGroupsService) Trigger(ctx context.Context, orderGroupID string, opts *types.MarginOrderGroupOpts) error {
	return s.client.put(ctx, joinPath("margin", "order_groups", orderGroupID, "trigger"), perpsOrderGroupQuery(opts), nil, nil)
}

// UpdateLimit changes the rolling 15-second contracts limit. A limit already
// exceeded triggers the group immediately.
func (s *PerpsOrderGroupsService) UpdateLimit(ctx context.Context, orderGroupID string, req *types.UpdateOrderGroupLimitRequest, opts *types.MarginOrderGroupOpts) error {
	if req == nil || (req.ContractsLimit == nil && req.ContractsLimitFp == nil) {
		return errors.New("contracts_limit or contracts_limit_fp required")
	}
	return s.client.put(ctx, joinPath("margin", "order_groups", orderGroupID, "limit"), perpsOrderGroupQuery(opts), req, nil)
}

func perpsOrderGroupQuery(opts *types.MarginOrderGroupOpts) url.Values {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	return v
}
