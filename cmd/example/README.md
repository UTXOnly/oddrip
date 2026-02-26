# Example

Runs a large set of **read-only** Kalshi REST API calls (no orders placed, no money spent). All output is written to **`api_calls.log`** in the current directory; nothing is printed to stdout.

**Note:** The demo API (`demo-api.kalshi.co`) is not fully functional; many endpoints may return errors or empty data. That is expected. Use `BASE_URL` to point at production for a full run.

Each log entry includes:
- A short description of the call (e.g. `Exchange.GetStatus`, `Markets.List (limit=5)`)
- The HTTP method and **fully constructed URL** (base URL + path + query)
- Response status code and **raw response body** (pretty-printed JSON)

Endpoints covered: exchange (status, announcements, schedule, user_data_timestamp, historical cutoff, series fee changes), markets (list, get, orderbook, trades), orders (list, get, queue position(s)), portfolio (balance, fills, positions), account (API limits).

## Keys from files in this directory

Place your Kalshi API credentials as files in this directory:

| File              | Description                          |
|-------------------|--------------------------------------|
| `key_id`          | Your API key ID (single line).        |
| `private_key.pem` | RSA private key in PEM format (PKCS#8 or PKCS#1). |

Then run from the repo root:

```bash
cd cmd/example
KALSHI_ACCESS_KEY=$(cat key_id) KALSHI_PRIVATE_KEY_PATH=./private_key.pem go run .
```

Or from anywhere, with paths to the key files:

```bash
KALSHI_ACCESS_KEY=$(cat /path/to/example/key_id) \
KALSHI_PRIVATE_KEY_PATH=/path/to/example/private_key.pem \
go run ./cmd/example
```

After running, open `api_calls.log` in the same directory to see the request URLs and raw responses.

## Environment variables

| Variable                 | Required | Description |
|--------------------------|----------|-------------|
| `KALSHI_ACCESS_KEY`      | For auth | API key ID. |
| `KALSHI_PRIVATE_KEY_PATH`| For auth | Path to PEM file (RSA private key). Used for request signing. |
| `BASE_URL`               | No       | API base URL. **Default: demo** `https://demo-api.kalshi.co/trade-api/v2`. Set to `https://api.elections.kalshi.com/trade-api/v2` for production. |
| `LIVE`                   | No       | Set to `1` to run live order flow (find open 15m BTC market, place 1-contract yes limit at 1¢, then cancel). Use with **production** and auth. |

Without auth, only public endpoints run. With auth, portfolio and orders endpoints are called as well.

**Live mode (place order):** Set `LIVE=1` and use **production** `BASE_URL` (demo does not support order placement). The example will: (1) find the current open 15‑minute BTC market (series `KXBTC15M`), (2) place a limit order for 1 contract yes at 1¢ and leave it resting so you can confirm, (3) place a second 1¢ yes bid, (4) cancel only the second order (to test cancel). The first order remains resting. Requires auth.

**Security:** Do not commit `key_id` or `private_key.pem`. Add them to `.gitignore` if they live under the repo.
