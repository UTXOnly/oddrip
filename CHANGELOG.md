# Changelog

All notable changes to this project are documented here. The client tracks [Kalshi’s API changelog](https://docs.kalshi.com/changelog); repository root `openapi.yaml` / `asyncapi.yaml` are the source of truth for shapes and endpoints.

## [Unreleased]

## [0.2.0] — 2026-03-21

### Added

- **Portfolio:** `ListSettlements`, `ListHistoricalFills`, and `ListHistoricalOrders` for `GET /portfolio/settlements`, `GET /historical/fills`, and `GET /historical/orders`.
- **Markets:** `ListHistorical`, `GetHistorical`, `GetHistoricalTrades`, and `GetHistoricalCandlesticks` for historical market data and archived trades.
- **Types:** `Settlement`, `GetSettlementsOpts`, historical candlestick payloads (`GetMarketCandlesticksHistoricalResponse`, nested distributions), and additional `Market` / `Fill` / `Order` fields aligned with OpenAPI 3.10 (e.g. `yes_price_dollars` / `no_price_dollars` on fills, `taker_fees_dollars` / `maker_fees_dollars` on orders, settlement and lifecycle fields on markets).
- **`oddrip.Version`** constant (`0.2.0`) for the client module; use git tag **`v0.2.0`** when publishing.

### Changed

- **Types:** `Trade` and queue-position types adjusted for current spec (e.g. `Trade.created_time` and dollar price fields; queue position responses emphasize `queue_position_fp`).
- **Types:** `MarketPosition.last_updated_ts` is always unmarshaled when present (required in the published contract).
