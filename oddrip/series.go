package oddrip

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type SeriesService struct {
	client *Client
}

func (s *SeriesService) List(ctx context.Context, opts *types.GetSeriesListOpts) (*types.GetSeriesListResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "category", opts.Category)
		encodeQuery(v, "tags", opts.Tags)
		encodeQueryBool(v, "include_product_metadata", opts.IncludeProductMetadata)
		encodeQueryBool(v, "include_volume", opts.IncludeVolume)
		encodeQueryInt64(v, "min_updated_ts", opts.MinUpdatedTs)
	}
	var out types.GetSeriesListResponse
	if err := s.client.get(ctx, joinPath("series"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SeriesService) Get(ctx context.Context, seriesTicker string, opts *types.GetSeriesOpts) (*types.GetSeriesResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryBool(v, "include_volume", opts.IncludeVolume)
	}
	var out types.GetSeriesResponse
	if err := s.client.get(ctx, joinPath("series", seriesTicker), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SeriesService) GetMarketCandlesticks(ctx context.Context, seriesTicker, ticker string, opts *types.GetMarketCandlesticksOpts) (*types.GetMarketCandlesticksResponse, error) {
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
	var out types.GetMarketCandlesticksResponse
	if err := s.client.get(ctx, joinPath("series", seriesTicker, "markets", ticker, "candlesticks"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SeriesService) GetEventCandlesticks(ctx context.Context, seriesTicker, eventTicker string, opts *types.GetEventCandlesticksOpts) (*types.GetEventCandlesticksResponse, error) {
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
	var out types.GetEventCandlesticksResponse
	if err := s.client.get(ctx, joinPath("series", seriesTicker, "events", eventTicker, "candlesticks"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SeriesService) GetForecastPercentileHistory(ctx context.Context, seriesTicker, eventTicker string, opts *types.GetEventForecastPercentilesHistoryOpts) (*types.GetEventForecastPercentilesHistoryResponse, error) {
	if opts == nil || len(opts.Percentiles) == 0 {
		return nil, errors.New("percentiles required")
	}
	if len(opts.Percentiles) > 10 {
		return nil, errors.New("at most 10 percentiles allowed")
	}
	switch opts.PeriodInterval {
	case 0, 1, 60, 1440:
	default:
		return nil, errors.New("period_interval must be 0, 1, 60, or 1440")
	}
	v := url.Values{}
	for _, p := range opts.Percentiles {
		v.Add("percentiles", strconv.Itoa(p))
	}
	v.Set("start_ts", fmt.Sprintf("%d", opts.StartTs))
	v.Set("end_ts", fmt.Sprintf("%d", opts.EndTs))
	v.Set("period_interval", fmt.Sprintf("%d", opts.PeriodInterval))
	var out types.GetEventForecastPercentilesHistoryResponse
	if err := s.client.get(ctx, joinPath("series", seriesTicker, "events", eventTicker, "forecast_percentile_history"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
