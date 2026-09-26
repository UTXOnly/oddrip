package types

// Types shared by more than one perps (margin) area. Area-specific perps types
// live in perps_*.go. ExchangeInstanceMargined and the book-side, time-in-force,
// and self-trade-prevention constants in common.go apply to perps as well.

// TickerPrice is a mark or reference price with its source timestamp, as
// carried by perps markets and the perps ticker channel.
type TickerPrice struct {
	Price string `json:"price"`
	TsMs  int64  `json:"ts_ms"`
}

// Perps market status values (MarginMarket.Status).
const (
	MarginMarketStatusInactive = "inactive"
	MarginMarketStatusActive   = "active"
	MarginMarketStatusClosed   = "closed"
)

// Perps order source and reason values. OrderSourceSystem marks orders the
// exchange placed for the user (liquidations, exit triggers); OrderReason says
// which.
const (
	OrderSourceUser   = "user"
	OrderSourceSystem = "system"

	OrderReasonLiquidation        = "liquidation"
	OrderReasonTakeProfitStopLoss = "take_profit_stop_loss"
)

// Perps order last-update reasons (MarginOrder.LastUpdateReason and the
// perps orderbook_delta and user_orders channels). The REST field is empty
// when no reason applies. CloseCancel and HaltCancel appear only in the
// WebSocket spec and are reserved there.
const (
	MarginLastUpdateReasonDecrease            = "Decrease"
	MarginLastUpdateReasonAmend               = "Amend"
	MarginLastUpdateReasonMarginCancel        = "MarginCancel"
	MarginLastUpdateReasonSelfTradeCancel     = "SelfTradeCancel"
	MarginLastUpdateReasonExpiryCancel        = "ExpiryCancel"
	MarginLastUpdateReasonCloseCancel         = "CloseCancel"
	MarginLastUpdateReasonHaltCancel          = "HaltCancel"
	MarginLastUpdateReasonTrade               = "Trade"
	MarginLastUpdateReasonPostOnlyCrossCancel = "PostOnlyCrossCancel"
	MarginLastUpdateReasonReduceOnlyCancel    = "ReduceOnlyCancel"
)
