# WebSocket example

Subscribe / list / update / unsubscribe / receive. No orders. Writes `ws_calls.log` in the current directory.

Uses the same credentials as the REST example (`KALSHI_ACCESS_KEY`, `KALSHI_PRIVATE_KEY_PATH`). Default is demo (`wss://demo-api.kalshi.co/trade-api/ws/v2`). Demo often returns `websocket: bad handshake`; set `BASE_URL=https://external-api.kalshi.com/trade-api/v2` for production. The WS URL is derived from `BASE_URL`: `external-api.kalshi.com` maps to `external-api-ws.kalshi.com`; any other host (the shared `api.elections.kalshi.com`, demo) is used as-is.

Exercises ticker, orderbook_delta, trade, market_lifecycle_v2, multi-market ticker with `send_initial_snapshot`, `ListSubscriptions`, `add_markets` / `delete_markets`. Does not call `get_snapshot`, Pyth, or CF Benchmarks.

```bash
cd cmd/example/websocket_example
KALSHI_ACCESS_KEY=$(cat ../key_id) KALSHI_PRIVATE_KEY_PATH=../private_key.pem go run .
```
