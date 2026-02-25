package types

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
	Service string `json:"service"`
}

type CursorPage struct {
	Cursor string `json:"cursor"`
}

const (
	OrderStatusResting   = "resting"
	OrderStatusCanceled  = "canceled"
	OrderStatusExecuted  = "executed"
	OrderSideYes         = "yes"
	OrderSideNo          = "no"
	OrderActionBuy       = "buy"
	OrderActionSell      = "sell"
	OrderTypeLimit       = "limit"
	OrderTypeMarket      = "market"
	TimeInForceFOK       = "fill_or_kill"
	TimeInForceGTC       = "good_till_canceled"
	TimeInForceIOC       = "immediate_or_cancel"
	SelfTradeTakerAtCross = "taker_at_cross"
	SelfTradeMaker        = "maker"
)

const (
	PeriodInterval1Min   = 1
	PeriodInterval1Hour  = 60
	PeriodInterval1Day   = 1440
)

const (
	MarketStatusUnopened  = "unopened"
	MarketStatusOpen      = "open"
	MarketStatusPaused    = "paused"
	MarketStatusClosed    = "closed"
	MarketStatusSettled   = "settled"
)

const (
	EventStatusOpen    = "open"
	EventStatusClosed  = "closed"
	EventStatusSettled = "settled"
)
