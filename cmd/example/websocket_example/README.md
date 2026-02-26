# WebSocket example

Runs WebSocket client methods that **do not place orders** (subscribe, list, update subscription, unsubscribe, receive messages). All output is written to **`ws_calls.log`** in the current directory; nothing is printed to stdout.

**Note:** The demo API (`demo-api.kalshi.co`) is not fully functional; many endpoints may return errors or empty data. That is expected. Set `BASE_URL` to production for a more complete run.

Each log entry includes:
- A short description of the call (e.g. `Subscribe (ticker, single market)`, `ListSubscriptions`)
- The **sent** payload (JSON command)
- The **response** (raw JSON) or error

Uses the same keys as the REST example: `KALSHI_ACCESS_KEY` and `KALSHI_PRIVATE_KEY_PATH` (path to PEM file). **Default: demo** — connects to `wss://demo-api.kalshi.co/trade-api/ws/v2`. Set `BASE_URL` to `https://api.elections.kalshi.com/trade-api/v2` to use production (WS URL is derived from `BASE_URL`).

**Methods exercised:** Connect (wss URL logged), Subscribe (ticker, orderbook_delta, trade, market_lifecycle_v2, ticker with multiple markets + send_initial_snapshot), ListSubscriptions, UpdateSubscription (add_markets, delete_markets), Receive messages (5s of ticker/orderbook/etc.), Unsubscribe, Close.

## Run

From repo root:

```bash
cd cmd/example/websocket_example
KALSHI_ACCESS_KEY=$(cat ../key_id) KALSHI_PRIVATE_KEY_PATH=../private_key.pem go run .
```

Or from `cmd/example`:

```bash
cd websocket_example
KALSHI_ACCESS_KEY=$(cat ../key_id) KALSHI_PRIVATE_KEY_PATH=../private_key.pem go run .
```

Then open `ws_calls.log` in the same directory.
