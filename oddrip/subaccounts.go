package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type SubaccountsService struct {
	client *Client
}

// Create adds a numbered subaccount (1-63). A nil req uses exchange index 0.
func (s *SubaccountsService) Create(ctx context.Context, req *types.CreateSubaccountRequest) (*types.CreateSubaccountResponse, error) {
	if req == nil {
		req = &types.CreateSubaccountRequest{}
	}
	var out types.CreateSubaccountResponse
	if err := s.client.post(ctx, joinPath("portfolio", "subaccounts"), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SubaccountsService) GetBalances(ctx context.Context) (*types.GetSubaccountBalancesResponse, error) {
	var out types.GetSubaccountBalancesResponse
	if err := s.client.get(ctx, joinPath("portfolio", "subaccounts", "balances"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SubaccountsService) Transfer(ctx context.Context, req *types.ApplySubaccountTransferRequest) error {
	if req == nil || req.ClientTransferID == "" {
		return errors.New("client_transfer_id required")
	}
	return s.client.post(ctx, joinPath("portfolio", "subaccounts", "transfer"), req, nil)
}

func (s *SubaccountsService) ListTransfers(ctx context.Context, opts *types.GetSubaccountTransfersOpts) (*types.GetSubaccountTransfersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetSubaccountTransfersResponse
	if err := s.client.get(ctx, joinPath("portfolio", "subaccounts", "transfers"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SubaccountsService) GetNetting(ctx context.Context) (*types.GetSubaccountNettingResponse, error) {
	var out types.GetSubaccountNettingResponse
	if err := s.client.get(ctx, joinPath("portfolio", "subaccounts", "netting"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SubaccountsService) UpdateNetting(ctx context.Context, req *types.UpdateSubaccountNettingRequest) error {
	if req == nil {
		return errors.New("request required")
	}
	return s.client.put(ctx, joinPath("portfolio", "subaccounts", "netting"), req, nil)
}
