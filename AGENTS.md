# AGENTS.md

Guidance for coding agents working in this repository. Humans: the README covers usage; this file covers how changes are made and verified.

## What this is

`github.com/UTXOnly/oddrip` is a Go client for the Kalshi Trade API (REST + WebSocket). Go 1.24+. Public module — do not rename it, and keep everything you commit tool-neutral and free of credentials, private project names, and personal trading setups.

```
oddrip/                 client, services (one file per service), ws.go, errors.go, retry glue
oddrip/types/           request/response structs, WS payloads, Dollars/Count/ParseTime helpers
oddrip/internal/auth    RSA-PSS signer (timestamp + METHOD + path, query excluded)
oddrip/internal/retry   backoff, Retry-After, idempotency-aware retry loop
cmd/example/            REST and WebSocket example programs
openapi.yaml            vendored Kalshi OpenAPI (REST)  — source of truth for shapes and paths
asyncapi.yaml           vendored Kalshi AsyncAPI (WS)   — source of truth for channels and messages
CHANGELOG.md            Keep a Changelog style; the release job extracts notes from it
```

## Source of truth

The vendored specs at the repo root define what the client should do. When they change, update `oddrip/types/`, the service methods, tests, `CHANGELOG.md`, and the version numbers named in the README intro. Published copies:

| Resource | URL |
|---|---|
| Changelog | https://docs.kalshi.com/changelog |
| OpenAPI | https://docs.kalshi.com/openapi.yaml |
| AsyncAPI | https://docs.kalshi.com/asyncapi.yaml |

To refresh: `curl -sL -o openapi.yaml https://docs.kalshi.com/openapi.yaml` (same for `asyncapi.yaml`), read `info.version` in each, then diff `paths:`, `components/schemas`, and AsyncAPI `channels` / `components/messages` against the Go code. Note the AsyncAPI keeps `info.version: 2.0.0` while its content changes; compare content, not just the version string.

Where the spec and production disagree, production wins, and the difference gets documented in the README. Known cases:

- Error bodies. The spec's `ErrorResponse` is flat `{"code","message","details"}`. Production returns `{"error":{"code":...,"message":...}}` for most errors and `{"msg":"..."}` for parameter-binding 400s. `newAPIError` parses all three; keep it that way if the spec changes. `APIError.RawBody` always has the body.
- Hosts. Spec primary REST host is `external-api.kalshi.com`; `api.elections.kalshi.com` is listed as also supported and is the client default. AsyncAPI names `external-api-ws.kalshi.com`; the client defaults to `api.elections.kalshi.com` for WS too. Both answer.
- `client_order_id` deduplication is documented in Kalshi's quick-start guide, not in the OpenAPI field description. A replay the server already applied returns 409.

## Conventions

- One `*Service` per API area on `Client`: `Exchange`, `Markets`, `Events`, `Series`, `Orders`, `OrderGroups`, `Portfolio`, `Subaccounts`, `Account`, `LiveData`. Every public method takes `context.Context` first.
- Build paths with `joinPath(...)`. It path-escapes each segment and returns `""` for an empty / `.` / `..` segment, which `do()` turns into `ErrEmptyPathParam` before sending. Never concatenate paths by hand; a dropped segment can route to a different endpoint (empty order ID → CancelAll).
- Optional query params go on a `*Opts` struct: `string` fields are sent when non-empty, numeric and bool fields are pointers. Use the `encodeQuery*` helpers. Array params: check the spec — `explode: true` arrays use `v.Add` per value (`tickers` on `/markets/orderbooks`, `percentiles`); comma-separated ones are a single `string` (`market_tickers` on `/markets/candlesticks`, `tickers` on `/markets`).
- Retry policy is chosen per call:
  - `get` / `put` / `delete` — idempotent: retried on 429, 5xx, and transport errors.
  - `postIdempotent` — only for POSTs the server deduplicates (`client_order_id`, `client_transfer_id`) or that set absolute state.
  - `post` / `postQuery` — everything else: retried on 429 only. A 5xx or dropped connection may mean the write landed.
  - `CreateV2` / `BatchCreateV2` pick at runtime based on whether every order has a `client_order_id`.
- Endpoints whose spec response is empty return `error` only.
- Prefer fixed-point `_fp` and `_dollars` string fields. Legacy integer price/count fields are being removed by Kalshi; do not add new ones.
- WebSocket: command replies are matched by `id`; `get_snapshot` is the exception (answered by `orderbook_snapshot` frames keyed by `sid`). Multi-channel `Subscribe` expects one `subscribed` reply per channel. All socket writes go through `writeMu`.
- Match the surrounding code. No drive-by refactors, reformatting, or comment rewrites in files you are not otherwise changing.

## Tests are mandatory

Every new or changed public REST method, `json`-tagged type, or WebSocket command/message shape gets a test in the same change.

| Change | Test file | Assert |
|---|---|---|
| REST method | `oddrip/*_test.go` (Series / OrderGroups / Subaccounts and extra Events / Markets / Portfolio methods live in `oddrip/services_extra_test.go`) | HTTP method, path, query params, request body when non-trivial. Use the `mockTransport` pattern from `oddrip/client_test.go`. |
| Request/response struct | `oddrip/types/*_test.go` | JSON unmarshal using the spec's example payload, not an invented one. Marshal too if the client constructs it. |
| WS channel, action, or payload | `oddrip/types/ws_test.go`, `oddrip/types/ws_messages_test.go`, `oddrip/ws_test.go` | Constant values, JSON round-trip from the AsyncAPI example, and — for commands that wait on a reply — that the call returns against a mock server. |

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
- Every claim in README and CHANGELOG about behavior must be true of the code as merged. If a fix is not in yet, describe the current behavior, not the intended one.
- Local review notes (`*-pre-release-review.md`, `*PLAN.md`) are gitignored. Do not commit them or reference them from public docs.
