# AGENTS.md

Guidance for coding agents working in this repository. Humans: the README covers usage; this file covers how changes are made and verified.

## What this is

`github.com/UTXOnly/oddrip` is a Go client for the Kalshi Trade API (REST + WebSocket). Go 1.24+. Public module — do not rename it, and keep everything you commit tool-neutral and free of credentials, private project names, and personal trading setups.

```
oddrip/                 client, services (one file per service), ws.go, errors.go, retry glue
oddrip/perps.go         PerpsClient (Client.Perps) and the perps service declarations
oddrip/perps_*.go       perps service methods, one file per perps area
oddrip/types/           request/response structs, WS payloads, Dollars/Count/ParseTime helpers
oddrip/types/perps*.go  perps (margin) structs, constants, and WS payloads
oddrip/internal/auth    RSA-PSS signer (timestamp + METHOD + path, query excluded)
oddrip/internal/retry   backoff, Retry-After, idempotency-aware retry loop
cmd/example/            REST and WebSocket example programs
openapi.yaml            vendored Kalshi OpenAPI (REST)  — source of truth for shapes and paths
asyncapi.yaml           vendored Kalshi AsyncAPI (WS)   — source of truth for channels and messages
perps_openapi.yaml      vendored Kalshi perps OpenAPI   — source of truth for perps REST
perps_asyncapi.yaml     vendored Kalshi perps AsyncAPI  — source of truth for perps WS
CHANGELOG.md            Keep a Changelog style; the release job extracts notes from it
```

## Source of truth

The vendored specs at the repo root define what the client should do. When they change, update `oddrip/types/`, the service methods, tests, `CHANGELOG.md`, and the version numbers named in the README intro. Published copies:

| Resource | URL |
|---|---|
| Changelog | https://docs.kalshi.com/changelog |
| OpenAPI | https://docs.kalshi.com/openapi.yaml |
| AsyncAPI | https://docs.kalshi.com/asyncapi.yaml |
| Perps OpenAPI | https://docs.kalshi.com/perps_openapi.yaml |
| Perps AsyncAPI | https://docs.kalshi.com/perps_asyncapi.yaml |
| Perps overview | https://docs.kalshi.com/margin.md |

To refresh: `curl -sL -o openapi.yaml https://docs.kalshi.com/openapi.yaml` (same for `asyncapi.yaml`), read `info.version` in each, then diff `paths:`, `components/schemas`, and AsyncAPI `channels` / `components/messages` against the Go code. Note the AsyncAPI keeps `info.version: 2.0.0` while its content changes; compare content, not just the version string.

Perps specs refresh the same way: `curl -sL -o perps_openapi.yaml https://docs.kalshi.com/perps_openapi.yaml` (same for `perps_asyncapi.yaml`). The perps OpenAPI currently reports `info.version: 0.0.1` (its title calls it "Manual Endpoints") and the perps AsyncAPI `2.0.0`, so again compare content, not versions. `/portfolio/intra_exchange_instance_transfer` appears in both REST specs; it is implemented once, on the event-contract `Portfolio` service (`CreateIntraExchangeTransfer`), not under `Perps`.

Where the spec and production disagree, production wins, and the difference gets documented in the README. Known cases:

- Error bodies. The spec's `ErrorResponse` is flat `{"code","message","details"}`. Production returns `{"error":{"code":...,"message":...}}` for most errors and `{"msg":"..."}` for parameter-binding 400s. `newAPIError` parses all three (from up to 64 KiB of the body); keep it that way if the spec changes. `APIError.RawBody` always has the first 512 bytes.
- Hosts. The client defaults to the spec's primary hosts: `external-api.kalshi.com` (REST) and `external-api-ws.kalshi.com` (WS). The shared `api.elections.kalshi.com` serves both protocols, is listed as also supported, and was the default before v0.6.2; `BaseURL` / `WSHost` select it. Perps REST uses the same base URL; the perps spec lists only `external-api.kalshi.com` and, for demo, `external-api.demo.kalshi.co` (not the shared `api.elections.kalshi.com` / `demo-api.kalshi.co`). The perps WebSocket has its own host, `external-api-margin-ws.kalshi.com` (demo `external-api-margin-ws.demo.kalshi.co`), path `/trade-api/ws/v2/margin`, reached via `Client.ConnectPerpsWS`.
- `client_order_id` deduplication is documented in Kalshi's quick-start guide, not in the OpenAPI field description. A replay the server already applied returns 409.

## Conventions

- One `*Service` per API area on `Client`: `Exchange`, `Markets`, `Events`, `Series`, `Orders`, `OrderGroups`, `Portfolio`, `Subaccounts`, `Account`, `LiveData`. Every public method takes `context.Context` first.
- Perps (Kalshi's perpetual futures; the API surface says "margin") hang off `client.Perps` (`*PerpsClient`), one sub-service per perps spec tag: `Exchange`, `Account`, `Markets`, `Orders`, `OrderGroups`, `Portfolio`, `Risk`, `Fees`, `Funding`, `ExitTriggers`, `FCM`. The service structs are declared in `oddrip/perps.go`; their methods live in per-area `oddrip/perps_<area>.go` files. Perps REST shares the `Client`'s base URL, signer, and retry policy, and every rule below applies to it. Perps POSTs keyed by `client_order_id` / `client_transfer_id` pick `postIdempotent` or `post` at runtime from whether the key is set, like `CreateV2`. `Perps.FCM.CreateSubtrader` always uses `postIdempotent`: the subtrader ID derives from the caller's suffix, so a replay gets 409 rather than a second subtrader.
- Build paths with `joinPath(...)`. It path-escapes each segment and returns `""` for an empty / `.` / `..` segment, which `do()` turns into `ErrEmptyPathParam` before sending. Never concatenate paths by hand; a dropped segment can route to a different endpoint (empty order ID → CancelAll).
- Optional query params go on a `*Opts` struct: `string` fields are sent when non-empty, numeric and bool fields are pointers. Use the `encodeQuery*` helpers. Array params: check the spec — `explode: true` arrays use `v.Add` per value (`tickers` on `/markets/orderbooks`, `percentiles`); comma-separated ones are a single `string` (`market_tickers` on `/markets/candlesticks`, `tickers` on `/markets`).
- Retry policy is chosen per call:
  - `get` / `put` / `delete` — idempotent: retried on 429, 5xx, and transport errors.
  - `postIdempotent` — only for POSTs the server deduplicates (`client_order_id`, `client_transfer_id`) or that set absolute state.
  - `post` / `postQuery` — everything else: retried on 429 only. A 5xx or dropped connection may mean the write landed.
  - `CreateV2` / `BatchCreateV2` pick at runtime based on whether every order has a `client_order_id`.
- Endpoints whose spec response is empty return `error` only.
- Prefer fixed-point `_fp` and `_dollars` string fields. Legacy integer price/count fields are being removed by Kalshi; do not add new ones.
- WebSocket: command replies are matched by `id`; `get_snapshot` is the exception (answered by `orderbook_snapshot` frames keyed by `sid`). Multi-channel `Subscribe` expects one `subscribed` reply per channel. Data-frame writes go through `writeSem` (a 1-slot channel so waiters can honor their context) and are bounded by `WSWriteTimeout`; control frames use gorilla's `WriteControl`, which serializes itself.
- Perps type names in `oddrip/types`: (1) a spec schema whose name contains `Margin` keeps it (`MarginOrder`, `CreateMarginOrderRequest`); (2) any other schema keeps its spec name when no existing type has it (`ExitTrigger`, `NotionalRiskLimitResponse`); (3) on a collision, reuse the existing type only if it matches the perps schema field-for-field (same JSON names and Go types, no extra fields), otherwise insert `Margin` after the leading verb, or prefix it when there is none (`ExchangeStatus` → `MarginExchangeStatus`, `ApplySubaccountTransferRequest` → `ApplyMarginSubaccountTransferRequest`).
- Perps WebSocket: `Client.ConnectPerpsWS` returns the same `*WSConn` as `ConnectWS`, pointed at the perps host and path, so the WebSocket rules above apply. Payloads are `Margin*Msg` types with millisecond timestamps (`ts_ms`, `created_ts_ms`, ...) as the perps AsyncAPI defines them; `order_group_updates` reuses `OrderGroupUpdatesMsg`. Channel names and frame `type` strings are the same ones event contracts use, so `type` alone does not tell a perps frame from an event-contract one.
- Match the surrounding code. No drive-by refactors, reformatting, or comment rewrites in files you are not otherwise changing.

## Tests are mandatory

Every new or changed public REST method, `json`-tagged type, or WebSocket command/message shape gets a test in the same change.

| Change | Test file | Assert |
|---|---|---|
| REST method | `oddrip/*_test.go` (Series / OrderGroups / Subaccounts and extra Events / Markets / Portfolio methods live in `oddrip/services_extra_test.go`) | HTTP method, path, query params, request body when non-trivial. Use the `mockTransport` pattern from `oddrip/client_test.go`. |
| Request/response struct | `oddrip/types/*_test.go` | JSON unmarshal using the spec's example payload, not an invented one. Marshal too if the client constructs it. |
| WS channel, action, or payload | `oddrip/types/ws_test.go`, `oddrip/types/ws_messages_test.go`, `oddrip/ws_test.go` | Constant values, JSON round-trip from the AsyncAPI example, and — for commands that wait on a reply — that the call returns against a mock server. |
| Perps REST method | `oddrip/perps_<area>_test.go` | As for REST methods (`newCaptureClient` and its `assertRequest` / `assertQuery` / `assertBody` in `oddrip/services_extra_test.go` cover these), plus `ErrEmptyPathParam` for path params, validation errors for required opts, and a retry test when the retry policy depends on the request. |
| Perps request/response struct | `oddrip/types/perps_<area>_test.go` (shared perps types: `oddrip/types/perps_test.go`) | JSON unmarshal from the perps spec's example. It has few, so otherwise build the payload from the schema's own properties (required fields plus notable optional ones, `null` for nullable ones) and never add fields the schema lacks. Marshal request structs the client builds. |
| Perps WS channel, command, or payload | `oddrip/perps_ws_test.go`, `oddrip/types/perps_ws_test.go` | As for WS above, against `perps_asyncapi.yaml`. |

If something cannot be unit-tested (live server behavior), add the closest test possible and say so in the PR.

## Verify before finishing

From the repo root:

```bash
gofmt -l . && go vet ./... && go build ./... && go test -race -count=1 ./...
```

CI additionally runs `staticcheck`, `govulncheck`, `go mod tidy` (must be a no-op), and the test matrix on Go 1.24 and stable across Linux, macOS, and Windows.

## Versioning and releases

`oddrip.Version` in `oddrip/version.go` is the release. Merging to `main` tags `v<Version>` and publishes a GitHub Release when that tag does not exist yet. Tags are immutable once on proxy.golang.org.

CI's `version` job enforces:

- `oddrip/version.go`, the top numbered `## [X.Y.Z]` in `CHANGELOG.md`, and the `@vX.Y.Z` pin in `README.md` agree.
- If `v<Version>` is already tagged, a PR may only touch files outside `oddrip/`, `go.mod`, `go.sum`. Any code change on a released version must bump the version and add a CHANGELOG section. Docs-only PRs need no bump.
- Semver at v0: a minor release may break; breaking changes go first under `### Breaking` in that version's section with migration notes. `gorelease` runs against the previous tag and fails the build if it finds API-incompatible changes without that heading, or if a patch bump declares one.

An `## [Unreleased]` heading above the numbered sections is ignored by the version check and is the place to accumulate notes between releases.

## Documentation

- README documents user-facing behavior: retry policy, WS lifecycle, coverage, where spec and production differ. It is not a method catalog; pkg.go.dev is.
- Keep the "N of 96 OpenAPI paths" count and the not-implemented list accurate. Count unique path templates the client hits (`joinPath` calls in non-test service files) against `paths:` in `openapi.yaml`.
- The README keeps a second count, "N of 38 perps OpenAPI paths", counted the same way against `paths:` in `perps_openapi.yaml`. `/portfolio/intra_exchange_instance_transfer` is in both specs, so it counts toward both totals.
- Every claim in README and CHANGELOG about behavior must be true of the code as merged. If a fix is not in yet, describe the current behavior, not the intended one.
- Local review notes (`*-pre-release-review.md`, `*PLAN.md`) are gitignored. Do not commit them or reference them from public docs.
