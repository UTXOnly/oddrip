package oddrip

import "context"

const (
	defaultPerpsWSHost = "external-api-margin-ws.kalshi.com"
	defaultPerpsWSPath = "/trade-api/ws/v2/margin"
)

// ConnectPerpsWS dials the perps (margin) WebSocket at
// wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin and returns
// a *WSConn that behaves exactly like one from ConnectWS. opts are applied
// after the perps host and path, so WSHost / WSPath override them; for demo
// pass WSHost("external-api-margin-ws.demo.kalshi.co").
//
// Only the orderbook_delta, ticker, trade, fill, user_orders, and
// order_group_updates channels exist, markets are selected by ticker, and
// UpdateSubscription supports add_markets, delete_markets, and get_snapshot.
// Payloads decode into the types.Margin*Msg structs (see types/perps_ws.go).
func (c *Client) ConnectPerpsWS(ctx context.Context, opts ...WSOption) (*WSConn, error) {
	return c.ConnectWS(ctx, perpsWSOptions(opts)...)
}

// perpsWSOptions puts the perps host and path ahead of opts so caller options
// win.
func perpsWSOptions(opts []WSOption) []WSOption {
	return append([]WSOption{WSHost(defaultPerpsWSHost), WSPath(defaultPerpsWSPath)}, opts...)
}
