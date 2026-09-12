package types

const (
	FeeTypeQuadratic                   = "quadratic"
	FeeTypeQuadraticWithMakerFees      = "quadratic_with_maker_fees"
	FeeTypeQuadraticWithComboMakerFees = "quadratic_with_combo_maker_fees"
	FeeTypeFlat                        = "flat"
)

// Series is one series. Category is the primary discovery category;
// Categories is the full list, which is what the category filter on
// Series.List matches against.
type Series struct {
	Ticker                 string                 `json:"ticker"`
	Frequency              string                 `json:"frequency"`
	Title                  string                 `json:"title"`
	Category               string                 `json:"category"`
	Categories             []string               `json:"categories"`
	Tags                   []string               `json:"tags"`
	SettlementSources      []SettlementSource     `json:"settlement_sources"`
	ContractURL            string                 `json:"contract_url"`
	ContractTermsURL       string                 `json:"contract_terms_url"`
	ProductMetadata        map[string]interface{} `json:"product_metadata,omitempty"`
	FeeType                string                 `json:"fee_type"`
	FeeMultiplier          float64                `json:"fee_multiplier"`
	AdditionalProhibitions []string               `json:"additional_prohibitions"`
	VolumeFp               string                 `json:"volume_fp,omitempty"`
	LastUpdatedTs          string                 `json:"last_updated_ts,omitempty"`
	ExchangeIndex          int                    `json:"exchange_index,omitempty"`
}

type GetSeriesResponse struct {
	Series Series `json:"series"`
}

type GetSeriesListResponse struct {
	Series []Series `json:"series"`
}

type GetSeriesListOpts struct {
	Category               string
	Tags                   string
	IncludeProductMetadata *bool
	IncludeVolume          *bool
	MinUpdatedTs           *int64
}

type GetSeriesOpts struct {
	IncludeVolume *bool
}

type BidAskDistribution struct {
	OpenDollars  string `json:"open_dollars"`
	LowDollars   string `json:"low_dollars"`
	HighDollars  string `json:"high_dollars"`
	CloseDollars string `json:"close_dollars"`
}

type PriceDistribution struct {
	OpenDollars     *string `json:"open_dollars,omitempty"`
	LowDollars      *string `json:"low_dollars,omitempty"`
	HighDollars     *string `json:"high_dollars,omitempty"`
	CloseDollars    *string `json:"close_dollars,omitempty"`
	MeanDollars     *string `json:"mean_dollars,omitempty"`
	PreviousDollars *string `json:"previous_dollars,omitempty"`
	MinDollars      *string `json:"min_dollars,omitempty"`
	MaxDollars      *string `json:"max_dollars,omitempty"`
}

type MarketCandlestick struct {
	EndPeriodTs    int64              `json:"end_period_ts"`
	YesBid         BidAskDistribution `json:"yes_bid"`
	YesAsk         BidAskDistribution `json:"yes_ask"`
	Price          PriceDistribution  `json:"price"`
	VolumeFp       string             `json:"volume_fp"`
	OpenInterestFp string             `json:"open_interest_fp"`
}

type GetMarketCandlesticksResponse struct {
	Ticker       string              `json:"ticker"`
	Candlesticks []MarketCandlestick `json:"candlesticks"`
}

type GetMarketCandlesticksOpts struct {
	StartTs                  int64
	EndTs                    int64
	PeriodInterval           int
	IncludeLatestBeforeStart *bool
}

type MarketCandlesticksResponse struct {
	MarketTicker string              `json:"market_ticker"`
	Candlesticks []MarketCandlestick `json:"candlesticks"`
}

type BatchGetMarketCandlesticksResponse struct {
	Markets []MarketCandlesticksResponse `json:"markets"`
}

type BatchGetMarketCandlesticksOpts struct {
	MarketTickers            string
	StartTs                  int64
	EndTs                    int64
	PeriodInterval           int
	IncludeLatestBeforeStart *bool
}

type GetEventCandlesticksResponse struct {
	MarketTickers      []string              `json:"market_tickers"`
	MarketCandlesticks [][]MarketCandlestick `json:"market_candlesticks"`
	AdjustedEndTs      int64                 `json:"adjusted_end_ts"`
}

type GetEventCandlesticksOpts struct {
	StartTs        int64
	EndTs          int64
	PeriodInterval int
}

type PercentilePoint struct {
	Percentile           int     `json:"percentile"`
	RawNumericalForecast float64 `json:"raw_numerical_forecast"`
	NumericalForecast    float64 `json:"numerical_forecast"`
	FormattedForecast    string  `json:"formatted_forecast"`
}

type ForecastPercentilesPoint struct {
	EventTicker      string            `json:"event_ticker"`
	EndPeriodTs      int64             `json:"end_period_ts"`
	PeriodInterval   int               `json:"period_interval"`
	PercentilePoints []PercentilePoint `json:"percentile_points"`
}

type GetEventForecastPercentilesHistoryResponse struct {
	ForecastHistory []ForecastPercentilesPoint `json:"forecast_history"`
}

// GetEventForecastPercentilesHistoryOpts: Percentiles holds 1-10 values in
// 0-9999; PeriodInterval is 0 (5-second), 1, 60, or 1440.
type GetEventForecastPercentilesHistoryOpts struct {
	Percentiles    []int
	StartTs        int64
	EndTs          int64
	PeriodInterval int
}
