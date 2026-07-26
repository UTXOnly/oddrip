package types

type ExchangeIndexStatus struct {
	ExchangeIndex                int  `json:"exchange_index"`
	ExchangeActive               bool `json:"exchange_active"`
	TradingActive                bool `json:"trading_active"`
	IntraExchangeTransfersActive bool `json:"intra_exchange_transfers_active"`
}

type ExchangeStatus struct {
	ExchangeActive               bool                  `json:"exchange_active"`
	TradingActive                bool                  `json:"trading_active"`
	IntraExchangeTransfersActive *bool                 `json:"intra_exchange_transfers_active,omitempty"`
	ExchangeEstimatedResumeTime  *string               `json:"exchange_estimated_resume_time,omitempty"`
	ExchangeIndexStatuses        []ExchangeIndexStatus `json:"exchange_index_statuses,omitempty"`
}

type GetHistoricalCutoffResponse struct {
	MarketSettledTs                string `json:"market_settled_ts"`
	TradesCreatedTs                string `json:"trades_created_ts"`
	OrdersUpdatedTs                string `json:"orders_updated_ts"`
	MarketPositionsLastUpdatedTs   string `json:"market_positions_last_updated_ts,omitempty"`
}

type GetUserDataTimestampResponse struct {
	AsOfTime string `json:"as_of_time"`
}

type DailySchedule struct {
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
}

type WeeklySchedule struct {
	StartTime string          `json:"start_time"`
	EndTime   string          `json:"end_time"`
	Monday    []DailySchedule `json:"monday"`
	Tuesday   []DailySchedule `json:"tuesday"`
	Wednesday []DailySchedule `json:"wednesday"`
	Thursday  []DailySchedule `json:"thursday"`
	Friday    []DailySchedule `json:"friday"`
	Saturday  []DailySchedule `json:"saturday"`
	Sunday    []DailySchedule `json:"sunday"`
}

type MaintenanceWindow struct {
	StartDatetime string `json:"start_datetime"`
	EndDatetime   string `json:"end_datetime"`
}

type Schedule struct {
	StandardHours      []WeeklySchedule    `json:"standard_hours"`
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
}

type GetExchangeScheduleResponse struct {
	Schedule Schedule `json:"schedule"`
}

type SeriesFeeChange struct {
	ID            string  `json:"id"`
	SeriesTicker  string  `json:"series_ticker"`
	FeeType       string  `json:"fee_type"`
	FeeMultiplier float64 `json:"fee_multiplier"`
	ScheduledTs   string  `json:"scheduled_ts"`
}

type GetSeriesFeeChangesResponse struct {
	SeriesFeeChangeArr []SeriesFeeChange `json:"series_fee_change_arr"`
}
