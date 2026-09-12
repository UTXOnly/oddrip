package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type OrderGroupsService struct {
	client *Client
}

func (s *OrderGroupsService) List(ctx context.Context, opts *types.GetOrderGroupsOpts) (*types.GetOrderGroupsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetOrderGroupsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "order_groups"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrderGroupsService) Create(ctx context.Context, req *types.CreateOrderGroupRequest) (*types.CreateOrderGroupResponse, error) {
	if req == nil || (req.ContractsLimit == nil && req.ContractsLimitFp == nil) {
		return nil, errors.New("contracts_limit or contracts_limit_fp required")
	}
	var out types.CreateOrderGroupResponse
	if err := s.client.post(ctx, joinPath("portfolio", "order_groups", "create"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *OrderGroupsService) Get(ctx context.Context, orderGroupID string, opts *types.GetOrderGroupOpts) (*types.GetOrderGroupResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetOrderGroupResponse
	if err := s.client.get(ctx, joinPath("portfolio", "order_groups", orderGroupID), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes the order group and cancels every order in it.
func (s *OrderGroupsService) Delete(ctx context.Context, orderGroupID string, opts *types.OrderGroupOpts) error {
	return s.client.delete(ctx, joinPath("portfolio", "order_groups", orderGroupID), orderGroupQuery(opts), nil, nil)
}

// Reset zeroes the group's matched-contracts counter so new orders can be placed.
func (s *OrderGroupsService) Reset(ctx context.Context, orderGroupID string, opts *types.OrderGroupOpts) error {
	return s.client.put(ctx, joinPath("portfolio", "order_groups", orderGroupID, "reset"), orderGroupQuery(opts), nil, nil)
}

// Trigger cancels every order in the group and blocks new orders until Reset.
func (s *OrderGroupsService) Trigger(ctx context.Context, orderGroupID string, opts *types.OrderGroupOpts) error {
	return s.client.put(ctx, joinPath("portfolio", "order_groups", orderGroupID, "trigger"), orderGroupQuery(opts), nil, nil)
}

// UpdateLimit changes the rolling 15-second contracts limit. A limit already
// exceeded triggers the group immediately.
func (s *OrderGroupsService) UpdateLimit(ctx context.Context, orderGroupID string, req *types.UpdateOrderGroupLimitRequest, opts *types.OrderGroupOpts) error {
	if req == nil || (req.ContractsLimit == nil && req.ContractsLimitFp == nil) {
		return errors.New("contracts_limit or contracts_limit_fp required")
	}
	return s.client.put(ctx, joinPath("portfolio", "order_groups", orderGroupID, "limit"), orderGroupQuery(opts), req, nil)
}

func orderGroupQuery(opts *types.OrderGroupOpts) url.Values {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		encodeQueryInt(v, "exchange_index", opts.ExchangeIndex)
	}
	return v
}
