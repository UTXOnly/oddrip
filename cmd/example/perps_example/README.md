# Perps example

Read-only perpetual futures (perps) calls over REST and the perps WebSocket. It places no orders and makes no cancels, amends, transfers, or other writes. Writes `perps_calls.log` in the current directory.

The API calls the product "margin" (paths under `/margin`, `Margin*` types); the client exposes it as `client.Perps` and `client.ConnectPerpsWS`.

Default is demo (`https://external-api.demo.kalshi.co/trade-api/v2`, the demo server in the perps spec). Set `BASE_URL=https://external-api.kalshi.com/trade-api/v2` for production. Kalshi is rolling perps out in production member by member, so on an account without access the authenticated calls and the WebSocket may return errors. Every error is logged and the run continues.

## What it calls

In order, all under `client.Perps`:

1. Public REST: `Exchange.GetStatus`, `Markets.List`, then for one market `Markets.Get`, `Markets.GetOrderbook` (depth 10), `Markets.GetTrades` (limit 10), `Markets.GetCandlesticks` (last hour, 1-minute), `Funding.GetRateEstimate`, `Funding.GetHistoricalRates` (last 24 hours), and finally `Risk.GetParameters`. The market is `PERPS_TICKER` if set, otherwise the first `active` market from `Markets.List`; with neither, the per-market calls and the WebSocket are skipped.
2. Authenticated REST: `Exchange.GetEnabled`, `Account.GetAPILimits`, `Portfolio.GetBalance`, `Portfolio.GetPositions`, `Portfolio.GetFills` (limit 5), `Orders.List` (limit 5), `OrderGroups.List`, `Risk.Get`, `Risk.GetNotionalRiskLimit`, `Fees.GetTiers`, `Fees.GetTierRates`.
3. WebSocket (every perps connection needs auth): `ConnectPerpsWS`, one `Subscribe` to `ticker` (with `send_initial_snapshot`) and `orderbook_delta` for the market, then up to 20 messages or 30 seconds, whichever comes first, then `Close`. `ticker`, `orderbook_snapshot`, `orderbook_delta`, and `error` frames are decoded into `types.MarginTickerMsg`, `types.MarginOrderbookSnapshotMsg`, `types.MarginOrderbookDeltaMsg`, and `types.ErrorMsg`; anything else (such as the `subscribed` replies) is logged raw.

Without auth, only step 1 runs. Does not call `Funding.GetHistory`, `Orders.Get`, `OrderGroups.Get`, exit triggers, FCM, or any endpoint that writes.

The WebSocket host is derived from `BASE_URL`: a demo host (`*.kalshi.co`) maps to `external-api-margin-ws.demo.kalshi.co`, anything else to `external-api-margin-ws.kalshi.com`. The path is `/trade-api/ws/v2/margin` on both.

## Running

Uses the same credentials as the REST example (see [`../README.md`](../README.md)):

```bash
cd cmd/example/perps_example
KALSHI_ACCESS_KEY=$(cat ../key_id) KALSHI_PRIVATE_KEY_PATH=../private_key.pem go run .
```

| Variable | Required | Description |
|----------|----------|-------------|
| `KALSHI_ACCESS_KEY` | for auth | API key ID |
| `KALSHI_PRIVATE_KEY_PATH` | for auth | Path to the PEM file |
| `BASE_URL` | no | Default: demo. Production: `https://external-api.kalshi.com/trade-api/v2` |
| `PERPS_TICKER` | no | Market for the per-market calls and the WebSocket subscription. Default: the first `active` market from `Markets.List` |
