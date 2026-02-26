package oddrip

import (
	"context"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type EventsService struct {
	client *Client
}

func (s *EventsService) List(ctx context.Context, opts *types.GetEventsOpts) (*types.GetEventsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQueryBool(v, "with_nested_markets", opts.WithNestedMarkets)
		encodeQueryBool(v, "with_milestones", opts.WithMilestones)
		encodeQuery(v, "status", opts.Status)
		encodeQuery(v, "series_ticker", opts.SeriesTicker)
		encodeQueryInt64(v, "min_close_ts", opts.MinCloseTs)
		encodeQueryInt64(v, "min_updated_ts", opts.MinUpdatedTs)
	}
	var out types.GetEventsResponse
	if err := s.client.get(ctx, joinPath("events"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *EventsService) ListMultivariate(ctx context.Context, opts *types.GetMultivariateEventsOpts) (*types.GetMultivariateEventsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
		encodeQuery(v, "series_ticker", opts.SeriesTicker)
		encodeQuery(v, "collection_ticker", opts.CollectionTicker)
		encodeQueryBool(v, "with_nested_markets", opts.WithNestedMarkets)
	}
	var out types.GetMultivariateEventsResponse
	if err := s.client.get(ctx, joinPath("events", "multivariate"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *EventsService) Get(ctx context.Context, eventTicker string, opts *types.GetEventOpts) (*types.GetEventResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryBool(v, "with_nested_markets", opts.WithNestedMarkets)
	}
	var out types.GetEventResponse
	if err := s.client.get(ctx, joinPath("events", eventTicker), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *EventsService) GetMetadata(ctx context.Context, eventTicker string) (*types.GetEventMetadataResponse, error) {
	var out types.GetEventMetadataResponse
	if err := s.client.get(ctx, joinPath("events", eventTicker, "metadata"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
