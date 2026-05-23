package oddrip

import (
	"context"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type AccountService struct {
	client *Client
}

func (s *AccountService) GetAPILimits(ctx context.Context) (*types.GetAccountApiLimitsResponse, error) {
	var out types.GetAccountApiLimitsResponse
	if err := s.client.get(ctx, joinPath("account", "limits"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AccountService) GetEndpointCosts(ctx context.Context) (*types.GetAccountEndpointCostsResponse, error) {
	var out types.GetAccountEndpointCostsResponse
	if err := s.client.get(ctx, joinPath("account", "endpoint_costs"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
