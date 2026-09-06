package oddrip

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type PortfolioService struct {
	client *Client
}

func (s *PortfolioService) GetBalance(ctx context.Context, opts *types.GetBalanceOpts) (*types.GetBalanceResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		encodeQueryInt(v, "exchange_index", opts.ExchangeIndex)
	}
	var out types.GetBalanceResponse
	if err := s.client.get(ctx, joinPath("portfolio", "balance"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) GetFills(ctx context.Context, opts *types.GetFillsOpts) (*types.GetFillsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "order_id", opts.OrderID)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetFillsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "fills"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) GetPositions(ctx context.Context, opts *types.GetPositionsOpts) (*types.GetPositionsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "count_filter", opts.CountFilter)
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		if opts.Limit != nil {
			v.Set("limit", fmt.Sprintf("%d", *opts.Limit))
		}
	}
	var out types.GetPositionsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "positions"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListSettlements(ctx context.Context, opts *types.GetSettlementsOpts) (*types.GetSettlementsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQueryInt64(v, "min_ts", opts.MinTs)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	var out types.GetSettlementsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "settlements"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListHistoricalFills(ctx context.Context, opts *types.GetHistoricalArchiveOpts) (*types.GetFillsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetFillsResponse
	if err := s.client.get(ctx, joinPath("historical", "fills"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListHistoricalOrders(ctx context.Context, opts *types.GetHistoricalArchiveOpts) (*types.GetOrdersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQueryInt64(v, "max_ts", opts.MaxTs)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetOrdersResponse
	if err := s.client.get(ctx, joinPath("historical", "orders"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListHistoricalPositions(ctx context.Context, opts *types.GetHistoricalPositionsOpts) (*types.GetPositionsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "ticker", opts.Ticker)
		encodeQuery(v, "event_ticker", opts.EventTicker)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetPositionsResponse
	if err := s.client.get(ctx, joinPath("historical", "positions"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListDeposits(ctx context.Context, opts *types.GetDepositsOpts) (*types.GetDepositsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetDepositsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "deposits"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PortfolioService) ListWithdrawals(ctx context.Context, opts *types.GetWithdrawalsOpts) (*types.GetWithdrawalsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetWithdrawalsResponse
	if err := s.client.get(ctx, joinPath("portfolio", "withdrawals"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListIntraExchangeTransfers returns intra-exchange account transfer history.
func (s *PortfolioService) ListIntraExchangeTransfers(ctx context.Context, opts *types.GetIntraExchangeTransfersOpts) (*types.GetIntraExchangeTransfersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetIntraExchangeTransfersResponse
	if err := s.client.get(ctx, joinPath("portfolio", "intra_exchange_instance_transfers"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIntraExchangeTransfer returns a single intra-exchange transfer by id.
func (s *PortfolioService) GetIntraExchangeTransfer(ctx context.Context, transferID string) (*types.GetIntraExchangeTransferResponse, error) {
	var out types.GetIntraExchangeTransferResponse
	if err := s.client.get(ctx, joinPath("portfolio", "intra_exchange_instance_transfers", transferID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTargetBalanceAllocation returns the caller's target balance allocation
// across exchange indexes.
func (s *PortfolioService) GetTargetBalanceAllocation(ctx context.Context) (*types.GetTargetBalanceAllocationResponse, error) {
	var out types.GetTargetBalanceAllocationResponse
	if err := s.client.get(ctx, joinPath("portfolio", "target_balance_allocation"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetTargetBalanceAllocation replaces the caller's target balance allocation.
// Percentages must total 100; an empty Allocations slice disables automatic
// rebalancing.
func (s *PortfolioService) SetTargetBalanceAllocation(ctx context.Context, req *types.SetTargetBalanceAllocationRequest) error {
	if req == nil {
		return errors.New("request required")
	}
	total := 0
	for _, a := range req.Allocations {
		total += a.Percent
	}
	if len(req.Allocations) > 0 && total != 100 {
		return fmt.Errorf("allocations must total 100, got %d", total)
	}
	return s.client.post(ctx, joinPath("portfolio", "target_balance_allocation"), req, nil)
}
