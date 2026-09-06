package oddrip

import (
	"context"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

type LiveDataService struct {
	client *Client
}

// GetEvent returns event-keyed live data (crypto price charts, commodity
// timeseries, weather observations) for an event ticker.
func (s *LiveDataService) GetEvent(ctx context.Context, eventTicker string, opts *types.GetEventLiveDataOpts) (*types.GetEventLiveDataResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "range", opts.Range)
	}
	var out types.GetEventLiveDataResponse
	if err := s.client.get(ctx, joinPath("live_data", "events", eventTicker), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWeatherIndex returns the Kalshi-computed city temperature index, the
// minute-resolution series behind hourly temperature markets. The window
// defaults to the last 24 hours; minutes where the index quorum failed are
// absent from the series.
func (s *LiveDataService) GetWeatherIndex(ctx context.Context, city string, opts *types.GetWeatherIndexOpts) (*types.GetWeatherIndexResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt64(v, "from", opts.From)
		encodeQueryInt64(v, "to", opts.To)
		encodeQueryInt64(v, "last_sec", opts.LastSec)
		encodeQueryBool(v, "detailed", opts.Detailed)
	}
	var out types.GetWeatherIndexResponse
	if err := s.client.get(ctx, joinPath("live_data", "weather", city), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWeatherIndexCalibrations returns a city's published index configuration
// timeline, ascending by effective time: the station weights, offsets, and city
// reference used to compute index values under each configuration version.
func (s *LiveDataService) GetWeatherIndexCalibrations(ctx context.Context, city string) (*types.GetWeatherIndexCalibrationsResponse, error) {
	var out types.GetWeatherIndexCalibrationsResponse
	if err := s.client.get(ctx, joinPath("live_data", "weather", city, "calibrations"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
