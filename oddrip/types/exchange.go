package types

type ExchangeStatus struct {
	ExchangeActive             bool    `json:"exchange_active"`
	TradingActive              bool    `json:"trading_active"`
	ExchangeEstimatedResumeTime *string `json:"exchange_estimated_resume_time,omitempty"`
}

type GetHistoricalCutoffResponse struct {
	MarketSettledTs  string `json:"market_settled_ts"`
	TradesCreatedTs  string `json:"trades_created_ts"`
	OrdersUpdatedTs  string `json:"orders_updated_ts"`
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
	StandardHours      []WeeklySchedule   `json:"standard_hours"`
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
}

type GetExchangeScheduleResponse struct {
	Schedule Schedule `json:"schedule"`
}

type Announcement struct {
	Type         string `json:"type"`
	Message      string `json:"message"`
	DeliveryTime string `json:"delivery_time"`
	Status       string `json:"status"`
}

type GetExchangeAnnouncementsResponse struct {
	Announcements []Announcement `json:"announcements"`
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
