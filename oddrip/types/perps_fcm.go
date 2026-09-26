package types

// Perps (margin) FCM member types: subtraders, their initial margin caps, and
// the member's own notional value risk limit.

// Asset classes for FCM initial margin caps (asset_class).
const (
	MarginAssetClassCrypto   = "Crypto"
	MarginAssetClassEquities = "Equities"
	MarginAssetClassMetals   = "Metals"
)

// CreateMarginFCMSubtraderRequest: SubtraderSuffix is 1-16 ASCII letters or
// digits. The server composes the subtrader ID as {user_id}_{SubtraderSuffix}.
type CreateMarginFCMSubtraderRequest struct {
	SubtraderSuffix string `json:"subtrader_suffix"`
}

type CreateMarginFCMSubtraderResponse struct {
	SubtraderID string `json:"subtrader_id"`
}

// GetFCMSubtraderRiskControlsOpts: SubtraderID is required unless the API key
// is bound to a subtrader, in which case it defaults to (and must equal) that
// subtrader. MarketTicker and AssetClass are mutually exclusive and filter
// RiskControls only.
type GetFCMSubtraderRiskControlsOpts struct {
	SubtraderID  string
	MarketTicker string
	AssetClass   string
}

// DeleteFCMSubtraderRiskControlsOpts: SubtraderID is required. MarketTicker or
// AssetClass (mutually exclusive) scope the removal to that cap.
type DeleteFCMSubtraderRiskControlsOpts struct {
	SubtraderID  string
	MarketTicker string
	AssetClass   string
}

// UpdateFCMSubtraderRiskControlsRequest sets an initial margin cap. ImCap is a
// non-negative fixed-point dollar string with up to 4 decimals. MarketTicker
// or AssetClass (mutually exclusive) scope the cap; with neither it applies
// across all markets.
type UpdateFCMSubtraderRiskControlsRequest struct {
	SubtraderID  string `json:"subtrader_id"`
	MarketTicker string `json:"market_ticker,omitempty"`
	AssetClass   string `json:"asset_class,omitempty"`
	ImCap        string `json:"im_cap"`
}

// GetFCMSubtraderRiskControlsResponse returns the FCM-set initial margin caps
// and the admin-set (read-only) notional value risk limits for one subtrader.
type GetFCMSubtraderRiskControlsResponse struct {
	RiskControls   []FCMSubtraderRiskControls      `json:"risk_controls"`
	NotionalLimits []FCMSubtraderNotionalRiskLimit `json:"notional_limits"`
}

// FCMSubtraderRiskControls is one initial margin cap. A cap with neither
// MarketTicker nor AssetClass applies across all markets. CurrentIm is the
// initial margin currently attributed to the cap's scope, from the exchange's
// read model, so it excludes in-flight orders and may trail the engine.
type FCMSubtraderRiskControls struct {
	SubtraderID  string `json:"subtrader_id"`
	MarketTicker string `json:"market_ticker,omitempty"`
	AssetClass   string `json:"asset_class,omitempty"`
	ImCap        string `json:"im_cap"`
	CurrentIm    string `json:"current_im"`
}

// FCMSubtraderNotionalRiskLimit is one admin-set notional value risk limit.
// An empty MarketTicker marks the whole-subtrader (all-markets) limit.
type FCMSubtraderNotionalRiskLimit struct {
	SubtraderID            string `json:"subtrader_id"`
	MarketTicker           string `json:"market_ticker,omitempty"`
	NotionalValueRiskLimit string `json:"notional_value_risk_limit"`
	CurrentNotional        string `json:"current_notional"`
}

// UpdateFCMNotionalRiskLimitRequest: NotionalValueRiskLimit is a non-negative
// fixed-point dollar string with up to 4 decimals.
type UpdateFCMNotionalRiskLimitRequest struct {
	NotionalValueRiskLimit string `json:"notional_value_risk_limit"`
}
