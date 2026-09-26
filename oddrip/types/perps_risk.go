package types

// Perps (margin) risk, fee, and funding types.

// GetMarginRiskParametersResponse holds the exchange-wide perps risk
// parameters. InitialMarginMultiplier maps market ticker to the factor applied
// to maintenance margin to get the initial margin requirement.
type GetMarginRiskParametersResponse struct {
	LiquidationMarginRatioThreshold float64            `json:"liquidation_margin_ratio_threshold"`
	QueueEntryMarginRatioThreshold  float64            `json:"queue_entry_margin_ratio_threshold"`
	InitialMarginMultiplier         map[string]float64 `json:"initial_margin_multiplier"`
}

// NotionalRiskLimitResponse holds the account's notional value risk limits and
// current usage, all as fixed-point dollar strings. A per-market limit
// overrides DefaultNotionalValueRiskLimit. MemberNotionalValueRiskLimit (FCM
// members only) and EffectiveAccountNotionalValueRiskLimit are empty when no
// such limit is set.
type NotionalRiskLimitResponse struct {
	DefaultNotionalValueRiskLimit          string            `json:"default_notional_value_risk_limit"`
	NotionalValueRiskLimitsByMarketTicker  map[string]string `json:"notional_value_risk_limits_by_market_ticker"`
	TotalCurrentUsage                      string            `json:"total_current_usage"`
	CurrentUsageByMarketTicker             map[string]string `json:"current_usage_by_market_ticker"`
	MemberNotionalValueRiskLimit           string            `json:"member_notional_value_risk_limit,omitempty"`
	EffectiveAccountNotionalValueRiskLimit string            `json:"effective_account_notional_value_risk_limit,omitempty"`
}

// MarginRiskPosition is one position's risk breakdown. The pointer fields are
// null when margin config is missing, when maintenance margin is zero, when no
// liquidation price exists, or when IsPortfolio is true (the position is
// hedged and cannot be attributed individually).
type MarginRiskPosition struct {
	Subaccount                int      `json:"subaccount"`
	MarketTicker              string   `json:"market_ticker"`
	Position                  string   `json:"position"`
	MarkPrice                 string   `json:"mark_price"`
	PositionNotional          string   `json:"position_notional"`
	MaintenanceMarginRequired *string  `json:"maintenance_margin_required,omitempty"`
	PositionLeverage          *float64 `json:"position_leverage,omitempty"`
	EstimatedLiquidationPrice *string  `json:"estimated_liquidation_price,omitempty"`
	IsPortfolio               bool     `json:"is_portfolio"`
}

// GetMarginRiskResponse: AccountLeverage is null when total maintenance margin
// is zero.
type GetMarginRiskResponse struct {
	AccountLeverage        *float64             `json:"account_leverage,omitempty"`
	TotalPositionNotional  string               `json:"total_position_notional"`
	TotalMaintenanceMargin string               `json:"total_maintenance_margin"`
	Positions              []MarginRiskPosition `json:"positions"`
}

// GetMarginFeeTiersResponse maps market ticker to the account's effective fee
// rate, as a decimal fraction of notional (0.0005 = 5 bps).
type GetMarginFeeTiersResponse struct {
	MakerFeeRates map[string]float64 `json:"maker_fee_rates"`
	TakerFeeRates map[string]float64 `json:"taker_fee_rates"`
}

// Perps fee schedules (MarginFeeTierRate.FeeSchedule).
const (
	MarginFeeScheduleSelfClearingMembers = "self_clearing_members"
	MarginFeeScheduleKalshiPrime         = "kalshi_prime"
	MarginFeeScheduleFCM                 = "fcm"
)

// MarginFeeTierRate is one tier of a fee schedule. Rates are decimal fractions
// of notional (0.0012 = 12 bps).
type MarginFeeTierRate struct {
	FeeSchedule  string  `json:"fee_schedule"`
	Tier         int     `json:"tier"`
	MakerFeeRate float64 `json:"maker_fee_rate"`
	TakerFeeRate float64 `json:"taker_fee_rate"`
}

// GetMarginFeeTierRatesResponse lists tiers from the entry tier upward.
type GetMarginFeeTierRatesResponse struct {
	FeeTierRates []MarginFeeTierRate `json:"fee_tier_rates"`
}

// MarginFundingHistoryEntry is one funding payment. FundingAmount is positive
// when received and negative when paid.
type MarginFundingHistoryEntry struct {
	MarketTicker     string  `json:"market_ticker"`
	FundingTime      string  `json:"funding_time"`
	FundingRate      float64 `json:"funding_rate"`
	MarkPrice        string  `json:"mark_price"`
	FundingAmount    string  `json:"funding_amount"`
	Quantity         string  `json:"quantity"`
	SubaccountNumber *int    `json:"subaccount_number,omitempty"`
}

type GetMarginFundingHistoryResponse struct {
	FundingHistory []MarginFundingHistoryEntry `json:"funding_history"`
}

type GetMarginFundingHistoryOpts struct {
	// Ticker limits results to one market; empty queries all markets.
	Ticker string
	// StartDate and EndDate are required, inclusive UTC dates in YYYY-MM-DD
	// form (e.g. "2026-01-31").
	StartDate string
	EndDate   string
	// Subaccount is 0 for primary, 1-63 for subaccounts; nil means all.
	Subaccount *int
}

type MarginFundingRate struct {
	MarketTicker string  `json:"market_ticker"`
	FundingTime  string  `json:"funding_time"`
	FundingRate  float64 `json:"funding_rate"`
	MarkPrice    string  `json:"mark_price"`
}

type GetMarginHistoricalFundingRatesResponse struct {
	FundingRates []MarginFundingRate `json:"funding_rates"`
}

type GetMarginHistoricalFundingRatesOpts struct {
	// Ticker limits results to one market; empty queries all markets.
	Ticker string
	// StartTs and EndTs are Unix seconds. The server defaults them to the
	// earliest available data and the current time.
	StartTs *int64
	EndTs   *int64
}

// GetMarginFundingRateEstimateResponse is the running estimate for the
// in-progress funding period; it keeps moving until NextFundingTime. Only
// NextFundingTime is required by the spec.
type GetMarginFundingRateEstimateResponse struct {
	MarketTicker    string   `json:"market_ticker,omitempty"`
	ComputedTime    string   `json:"computed_time,omitempty"`
	FundingRate     *float64 `json:"funding_rate,omitempty"`
	MarkPrice       string   `json:"mark_price,omitempty"`
	NextFundingTime string   `json:"next_funding_time"`
}

type GetMarginFundingRateEstimateOpts struct {
	// Ticker is required.
	Ticker string
}
