# REST example

Read-only Kalshi REST calls (no orders unless `LIVE=1`). Writes `api_calls.log` in the current directory.

Demo (`https://demo-api.kalshi.co/trade-api/v2`) is the default and is incomplete; many calls error or return empty data. Set `BASE_URL` to production for a real run.

Covers exchange, markets (list/get/orderbook/trades), events, orders (list/get/queue), portfolio (balance/fills/positions), and account limits. Does not call Series, OrderGroups, Subaccounts, LiveData, historical, or settlements.

## Credentials

Create two files in this directory (gitignored via `cmd/example/key_id` and `*.pem`):

| File | Contents |
|------|----------|
| `key_id` | API key ID, single line |
| `private_key.pem` | RSA private key, PKCS#8 or PKCS#1 PEM |

The program reads them through env vars, not those filenames:

```bash
cd cmd/example
KALSHI_ACCESS_KEY=$(cat key_id) KALSHI_PRIVATE_KEY_PATH=./private_key.pem go run .
```

| Variable | Required | Description |
|----------|----------|-------------|
| `KALSHI_ACCESS_KEY` | for auth | API key ID |
| `KALSHI_PRIVATE_KEY_PATH` | for auth | Path to the PEM file |
| `BASE_URL` | no | Default: demo. Production: `https://api.elections.kalshi.com/trade-api/v2` |
| `LIVE` | no | `1` places two 1¢ yes bids on the open 15m BTC market (`KXBTC15M`) and cancels the second. Uses production. Leaves the first order resting. |

Without auth, only public endpoints run.
