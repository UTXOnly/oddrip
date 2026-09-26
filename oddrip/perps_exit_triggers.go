package oddrip

import (
	"context"
	"errors"
	"net/url"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

// SetIsolated sets the bracket or trailing stop on the isolated position in
// ticker. It requires an open position and closes it in full when it fires.
func (s *PerpsExitTriggersService) SetIsolated(ctx context.Context, ticker string, req *types.SetIsolatedExitTriggerRequest) (*types.ExitTrigger, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	if !exitTriggerHasLeg(req.StopLossPrice, req.TakeProfitPrice, req.TrailAmount, req.TrailBps) {
		return nil, errExitTriggerNoLeg
	}
	var out types.ExitTrigger
	if err := s.client.put(ctx, joinPath("margin", "isolated", "positions", ticker, "exit_trigger"), nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIsolated lists the live exit triggers on the isolated position in ticker.
func (s *PerpsExitTriggersService) GetIsolated(ctx context.Context, ticker string, opts *types.IsolatedMarginExitTriggerOpts) (*types.GetExitTriggersResponse, error) {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "kind", opts.Kind)
	}
	var out types.GetExitTriggersResponse
	if err := s.client.get(ctx, joinPath("margin", "isolated", "positions", ticker, "exit_trigger"), v, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteIsolated cancels the exit triggers on the isolated position in ticker
// (all kinds unless opts.Kind is set). It succeeds when nothing is live and
// leaves the position itself open.
func (s *PerpsExitTriggersService) DeleteIsolated(ctx context.Context, ticker string, opts *types.IsolatedMarginExitTriggerOpts) error {
	v := url.Values{}
	if opts != nil {
		encodeQuery(v, "kind", opts.Kind)
	}
	return s.client.delete(ctx, joinPath("margin", "isolated", "positions", ticker, "exit_trigger"), v, nil, nil)
}

// SetCross sets a bracket or trailing stop on the cross (non-isolated)
// position in ticker. A position holds one trailing stop and up to 20
// brackets. Without Count or AnchorOrderID the write replaces the
// whole-position trigger of that kind; with either it appends a trigger and
// requires ClientTriggerID. Like every PUT it is retried on 5xx and transport
// errors: a replace is idempotent, and a replayed ClientTriggerID returns the
// trigger the first attempt created rather than adding another.
func (s *PerpsExitTriggersService) SetCross(ctx context.Context, ticker string, req *types.SetCrossExitTriggerRequest, opts *types.CrossMarginExitTriggerSubaccountOpts) (*types.ExitTrigger, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	if !exitTriggerHasLeg(req.StopLossPrice, req.TakeProfitPrice, req.TrailAmount, req.TrailBps) {
		return nil, errExitTriggerNoLeg
	}
	var out types.ExitTrigger
	if err := s.client.put(ctx, joinPath("margin", "cross", "positions", ticker, "exit_trigger"), crossExitTriggerSubaccountQuery(opts), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCross lists the live exit triggers on the cross position in ticker.
func (s *PerpsExitTriggersService) GetCross(ctx context.Context, ticker string, opts *types.CrossMarginExitTriggerOpts) (*types.GetExitTriggersResponse, error) {
	var out types.GetExitTriggersResponse
	if err := s.client.get(ctx, joinPath("margin", "cross", "positions", ticker, "exit_trigger"), crossExitTriggerQuery(opts), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCross cancels the exit triggers on the cross position in ticker (all
// kinds unless opts.Kind is set), including partial and order-anchored
// brackets. It succeeds when nothing is live. Use DeleteCrossByID to cancel a
// single bracket.
func (s *PerpsExitTriggersService) DeleteCross(ctx context.Context, ticker string, opts *types.CrossMarginExitTriggerOpts) error {
	return s.client.delete(ctx, joinPath("margin", "cross", "positions", ticker, "exit_trigger"), crossExitTriggerQuery(opts), nil, nil)
}

// UpdateCross replaces the legs of one bracket on the cross position in
// ticker, leaving other triggers untouched. Trailing stops have no ID route;
// use SetCross with Kind trailing.
func (s *PerpsExitTriggersService) UpdateCross(ctx context.Context, ticker, triggerID string, req *types.UpdateExitTriggerRequest, opts *types.CrossMarginExitTriggerSubaccountOpts) (*types.ExitTrigger, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	if !exitTriggerHasLeg(req.StopLossPrice, req.TakeProfitPrice, "", 0) {
		return nil, errors.New("stop_loss_price or take_profit_price required")
	}
	var out types.ExitTrigger
	if err := s.client.put(ctx, joinPath("margin", "cross", "positions", ticker, "exit_trigger", triggerID), crossExitTriggerSubaccountQuery(opts), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCrossByID cancels one bracket on the cross position in ticker,
// leaving other triggers live. Trailing stops have no ID route; use
// DeleteCross with Kind trailing.
func (s *PerpsExitTriggersService) DeleteCrossByID(ctx context.Context, ticker, triggerID string, opts *types.CrossMarginExitTriggerSubaccountOpts) error {
	return s.client.delete(ctx, joinPath("margin", "cross", "positions", ticker, "exit_trigger", triggerID), crossExitTriggerSubaccountQuery(opts), nil, nil)
}

// The server rejects a write that sets no leg; cancelling is a DELETE.
var errExitTriggerNoLeg = errors.New("stop_loss_price, take_profit_price, trail_amount, or trail_bps required")

func exitTriggerHasLeg(stopLoss, takeProfit, trailAmount string, trailBps int) bool {
	return stopLoss != "" || takeProfit != "" || trailAmount != "" || trailBps != 0
}

func crossExitTriggerQuery(opts *types.CrossMarginExitTriggerOpts) url.Values {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
		encodeQuery(v, "kind", opts.Kind)
	}
	return v
}

func crossExitTriggerSubaccountQuery(opts *types.CrossMarginExitTriggerSubaccountOpts) url.Values {
	v := url.Values{}
	if opts != nil {
		encodeQueryInt(v, "subaccount", opts.Subaccount)
	}
	return v
}
