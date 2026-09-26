package oddrip

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

func (s *PerpsExchangeService) GetStatus(ctx context.Context) (*types.MarginExchangeStatus, error) {
	var out types.MarginExchangeStatus
	if err := s.client.get(ctx, joinPath("margin", "exchange", "status"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEnabled reports whether perps trading is enabled for the authenticated
// user.
func (s *PerpsExchangeService) GetEnabled(ctx context.Context) (*types.MarginEnabledResponse, error) {
	var out types.MarginEnabledResponse
	if err := s.client.get(ctx, joinPath("margin", "enabled"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAPILimits returns the authenticated user's perps API rate-limit tier,
// which is separate from the event-contract tier returned by
// Client.Account.GetAPILimits.
func (s *PerpsAccountService) GetAPILimits(ctx context.Context) (*types.GetAccountApiLimitsResponse, error) {
	var out types.GetAccountApiLimitsResponse
	if err := s.client.get(ctx, joinPath("account", "limits", "perps"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsMarketsService) List(ctx context.Context, opts *types.GetMarginMarketsOpts) (*types.GetMarginMarketsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "status", opts.Status)
	}
	var out types.GetMarginMarketsResponse
	if err := s.client.get(ctx, joinPath("margin", "markets"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsMarketsService) Get(ctx context.Context, ticker string) (*types.MarginMarketResponse, error) {
	var out types.MarginMarketResponse
	if err := s.client.get(ctx, joinPath("margin", "markets", ticker), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsMarketsService) GetOrderbook(ctx context.Context, ticker string, opts *types.GetMarginMarketOrderbookOpts) (*types.MarginOrderbookResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "depth", opts.Depth)
		encodeQuery(v, "aggregation_tick_size", opts.AggregationTickSize)
	}
	var out types.MarginOrderbookResponse
	if err := s.client.get(ctx, joinPath("margin", "markets", ticker, "orderbook"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsMarketsService) GetCandlesticks(ctx context.Context, ticker string, opts *types.GetMarginMarketCandlesticksOpts) (*types.GetMarginMarketCandlesticksResponse, error) {
	if opts == nil {
		return nil, errors.New("opts required")
	}
	switch opts.PeriodInterval {
	case 1, 60, 1440:
	default:
		return nil, errors.New("period_interval must be 1, 60, or 1440")
	}
	v := url.Values{}
	v.Set("start_ts", fmt.Sprintf("%d", opts.StartTs))
	v.Set("end_ts", fmt.Sprintf("%d", opts.EndTs))
	v.Set("period_interval", fmt.Sprintf("%d", opts.PeriodInterval))
	encodeQueryBool(v, "include_latest_before_start", opts.IncludeLatestBeforeStart)
	var out types.GetMarginMarketCandlesticksResponse
	if err := s.client.get(ctx, joinPath("margin", "markets", ticker, "candlesticks"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PerpsMarketsService) GetTrades(ctx context.Context, opts *types.GetMarginTradesOpts) (*types.GetMarginTradesResponse, error) {
	if opts == nil || opts.Ticker == "" {
		return nil, errors.New("ticker required")
	}
	v := url.Values{}
	v.Set("ticker", opts.Ticker)
	encodeQueryInt64(v, "limit", opts.Limit)
	encodeQuery(v, "cursor", opts.Cursor)
	encodeQueryInt64(v, "min_ts", opts.MinTs)
	encodeQueryInt64(v, "max_ts", opts.MaxTs)
	var out types.GetMarginTradesResponse
	if err := s.client.get(ctx, joinPath("margin", "trades"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
