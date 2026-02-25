package oddrip

import (
	"context"
	"net/url"

	"github.com/oddrip/client/oddrip/types"
)

type ExchangeService struct {
	client *Client
}

func (s *ExchangeService) GetStatus(ctx context.Context) (*types.ExchangeStatus, error) {
	var out types.ExchangeStatus
	if err := s.client.get(ctx, joinPath("exchange", "status"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ExchangeService) GetAnnouncements(ctx context.Context) (*types.GetExchangeAnnouncementsResponse, error) {
	var out types.GetExchangeAnnouncementsResponse
	if err := s.client.get(ctx, joinPath("exchange", "announcements"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ExchangeService) GetSchedule(ctx context.Context) (*types.GetExchangeScheduleResponse, error) {
	var out types.GetExchangeScheduleResponse
	if err := s.client.get(ctx, joinPath("exchange", "schedule"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ExchangeService) GetUserDataTimestamp(ctx context.Context) (*types.GetUserDataTimestampResponse, error) {
	var out types.GetUserDataTimestampResponse
	if err := s.client.get(ctx, joinPath("exchange", "user_data_timestamp"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ExchangeService) GetHistoricalCutoff(ctx context.Context) (*types.GetHistoricalCutoffResponse, error) {
	var out types.GetHistoricalCutoffResponse
	if err := s.client.get(ctx, joinPath("historical", "cutoff"), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ExchangeService) GetSeriesFeeChanges(ctx context.Context, seriesTicker string, showHistorical bool) (*types.GetSeriesFeeChangesResponse, error) {
	v := url.Values{}
	if seriesTicker != "" {
		v.Set("series_ticker", seriesTicker)
	}
	if showHistorical {
		v.Set("show_historical", "true")
	}
	var out types.GetSeriesFeeChangesResponse
	if err := s.client.get(ctx, joinPath("series", "fee_changes"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
