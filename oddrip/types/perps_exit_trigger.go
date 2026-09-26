package types

// Perps (margin) exit trigger types: stop-loss / take-profit brackets and
// trailing stops that close a margin position with reduce-only orders.

// Exit trigger kinds (ExitTrigger.Kind, the request Kind fields, and the kind
// query parameter). Requests default to ExitTriggerKindBracket.
const (
	ExitTriggerKindBracket  = "bracket"
	ExitTriggerKindTrailing = "trailing"
)

// Exit trigger statuses (ExitTrigger.Status). Filled, failed, and canceled are
// terminal. An order-anchored trigger returns to PendingOnEntry after closing
// its current quantity, so that status can follow Active.
const (
	ExitTriggerStatusPendingOnEntry = "pending_on_entry"
	ExitTriggerStatusActive         = "active"
	ExitTriggerStatusFilled         = "filled"
	ExitTriggerStatusFailed         = "failed"
	ExitTriggerStatusCanceled       = "canceled"
	ExitTriggerStatusUnknown        = "unknown"
)

// Bracket legs (ExitTrigger.TriggeredLeg).
const (
	ExitTriggerLegStopLoss   = "stop_loss"
	ExitTriggerLegTakeProfit = "take_profit"
)

// Reasons a trigger reached its terminal status (ExitTrigger.StatusReason).
const (
	ExitTriggerReasonUserCanceled    = "user_canceled"
	ExitTriggerReasonPositionClosed  = "position_closed"
	ExitTriggerReasonEntryNotFilled  = "entry_not_filled"
	ExitTriggerReasonOrderRejected   = "order_rejected"
	ExitTriggerReasonPositionFlipped = "position_flipped"
)

// ExitTrigger is a bracket or trailing stop on a margin position. Prices are
// fixed-point dollar strings and counts fixed-point contract strings; a zero
// Count means the whole position. The bracket fields (StopLossPrice,
// TakeProfitPrice, TriggeredLeg) and the trailing fields (TrailAmount,
// TrailBps, WatermarkPrice, EffectiveStopPrice) are empty on the other kind.
// TriggeredOrderID survives an anchored trigger's return to pending_on_entry,
// and AnchorOrderID is cleared once the anchor order is terminal; correlate by
// ClientTriggerID instead.
type ExitTrigger struct {
	ID                 string `json:"id"`
	Ticker             string `json:"ticker"`
	Kind               string `json:"kind"`
	Status             string `json:"status"`
	StatusReason       string `json:"status_reason,omitempty"`
	TriggeredLeg       string `json:"triggered_leg,omitempty"`
	TriggeredOrderID   string `json:"triggered_order_id,omitempty"`
	AnchorOrderID      string `json:"anchor_order_id,omitempty"`
	ClientTriggerID    string `json:"client_trigger_id,omitempty"`
	StopLossPrice      string `json:"stop_loss_price,omitempty"`
	TakeProfitPrice    string `json:"take_profit_price,omitempty"`
	TrailAmount        string `json:"trail_amount,omitempty"`
	TrailBps           int    `json:"trail_bps,omitempty"`
	WatermarkPrice     string `json:"watermark_price,omitempty"`
	EffectiveStopPrice string `json:"effective_stop_price,omitempty"`
	Count              string `json:"count"`
	FilledCount        string `json:"filled_count"`
	CreatedTime        string `json:"created_time"`
	UpdatedTime        string `json:"updated_time"`
}

// GetExitTriggersResponse lists the live triggers on one position slot.
type GetExitTriggersResponse struct {
	ExitTriggers []ExitTrigger `json:"exit_triggers"`
}

// SetIsolatedExitTriggerRequest sets the trigger on an isolated position. For
// a bracket set StopLossPrice and/or TakeProfitPrice; for a trailing stop set
// exactly one of TrailAmount or TrailBps (1-9999). Fields of the other kind
// are rejected, as is a body that sets neither. A leg left empty is cleared.
type SetIsolatedExitTriggerRequest struct {
	Kind            string `json:"kind,omitempty"`
	StopLossPrice   string `json:"stop_loss_price,omitempty"`
	TakeProfitPrice string `json:"take_profit_price,omitempty"`
	TrailAmount     string `json:"trail_amount,omitempty"`
	TrailBps        int    `json:"trail_bps,omitempty"`
}

// SetCrossExitTriggerRequest sets or appends a trigger on a cross (non-isolated)
// position; the price and trail fields follow SetIsolatedExitTriggerRequest.
// Empty Count covers the whole position and replaces the existing trigger. A
// whole-number Count, or AnchorOrderID (bracket only, mutually exclusive with
// Count), appends a trigger instead and then requires ClientTriggerID, which
// is rejected on every other write. A replayed ClientTriggerID returns the
// trigger the first request created, in its current state.
type SetCrossExitTriggerRequest struct {
	Count           string `json:"count,omitempty"`
	AnchorOrderID   string `json:"anchor_order_id,omitempty"`
	ClientTriggerID string `json:"client_trigger_id,omitempty"`
	Kind            string `json:"kind,omitempty"`
	StopLossPrice   string `json:"stop_loss_price,omitempty"`
	TakeProfitPrice string `json:"take_profit_price,omitempty"`
	TrailAmount     string `json:"trail_amount,omitempty"`
	TrailBps        int    `json:"trail_bps,omitempty"`
}

// UpdateExitTriggerRequest replaces one bracket's legs. At least one leg is
// required, and a leg left empty is cleared.
type UpdateExitTriggerRequest struct {
	StopLossPrice   string `json:"stop_loss_price,omitempty"`
	TakeProfitPrice string `json:"take_profit_price,omitempty"`
}

// IsolatedMarginExitTriggerOpts carries the query parameters for
// PerpsExitTriggersService.GetIsolated and DeleteIsolated. An empty Kind
// covers every kind.
type IsolatedMarginExitTriggerOpts struct {
	Kind string
}

// CrossMarginExitTriggerOpts carries the query parameters for
// PerpsExitTriggersService.GetCross and DeleteCross. Subaccount defaults to 0
// (primary) server-side; an empty Kind covers every kind.
type CrossMarginExitTriggerOpts struct {
	Subaccount *int
	Kind       string
}

// CrossMarginExitTriggerSubaccountOpts carries the query parameter for
// PerpsExitTriggersService.SetCross, UpdateCross, and DeleteCrossByID.
// Subaccount defaults to 0 (primary) server-side.
type CrossMarginExitTriggerSubaccountOpts struct {
	Subaccount *int
}
