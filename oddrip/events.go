package oddrip

import (
	"context"
	"errors"
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
		encodeQuery(v, "tickers", opts.Tickers)
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

func (s *EventsService) ListMultivariateCollections(ctx context.Context, opts *types.GetMultivariateEventCollectionsOpts) (*types.GetMultivariateEventCollectionsResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "status", opts.Status)
		encodeQuery(v, "associated_event_ticker", opts.AssociatedEventTicker)
		encodeQuery(v, "series_ticker", opts.SeriesTicker)
		encodeQueryInt64(v, "limit", opts.Limit)
		encodeQuery(v, "cursor", opts.Cursor)
	}
	var out types.GetMultivariateEventCollectionsResponse
	if err := s.client.get(ctx, joinPath("multivariate_event_collections"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *EventsService) GetMultivariateCollection(ctx context.Context, collectionTicker string) (*types.GetMultivariateEventCollectionResponse, error) {
	var out types.GetMultivariateEventCollectionResponse
	if err := s.client.get(ctx, joinPath("multivariate_event_collections", collectionTicker), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateMarketInMultivariateCollection creates (or ensures) the combination
// market selected by req.SelectedMarkets. Must be called before trading or
// looking up such a market; limited to 5000 creations per week.
func (s *EventsService) CreateMarketInMultivariateCollection(ctx context.Context, collectionTicker string, req *types.CreateMarketInMultivariateEventCollectionRequest) (*types.CreateMarketInMultivariateEventCollectionResponse, error) {
	if req == nil || len(req.SelectedMarkets) == 0 {
		return nil, errors.New("selected_markets required")
	}
	var out types.CreateMarketInMultivariateEventCollectionResponse
	if err := s.client.post(ctx, joinPath("multivariate_event_collections", collectionTicker), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
