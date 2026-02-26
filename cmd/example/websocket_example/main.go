package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/oddrip/client/oddrip"
	"github.com/oddrip/client/oddrip/types"
)

const logFilename = "ws_calls.log"

const exampleMarketTicker = "KXBTC-24DEC31-T100000"

func main() {
	logFile, err := os.Create(logFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://demo-api.kalshi.co/trade-api/v2"
	}
	wsHost, wsPath := wsURLFromBase(baseURL)
	opts := []oddrip.Option{oddrip.BaseURL(baseURL)}
	if keyID := os.Getenv("KALSHI_ACCESS_KEY"); keyID != "" {
		if keyPath := os.Getenv("KALSHI_PRIVATE_KEY_PATH"); keyPath != "" {
			pemBytes, err := os.ReadFile(keyPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read private key: %v\n", err)
				os.Exit(1)
			}
			priv, err := oddrip.ParsePrivateKeyFromPEM(pemBytes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse private key: %v\n", err)
				os.Exit(1)
			}
			opts = append(opts, oddrip.Auth(oddrip.NewKalshiSigner(keyID, priv)))
		} else {
			sig := os.Getenv("KALSHI_ACCESS_SIGNATURE")
			ts := os.Getenv("KALSHI_ACCESS_TIMESTAMP")
			if sig != "" && ts != "" {
				opts = append(opts, oddrip.Auth(&oddrip.StaticHeaders{
					Headers: map[string][]string{
						"KALSHI-ACCESS-KEY":        {keyID},
						"KALSHI-ACCESS-SIGNATURE":  {sig},
						"KALSHI-ACCESS-TIMESTAMP":  {ts},
					},
				}))
			}
		}
	}

	client := oddrip.New(opts...)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	runAll(ctx, client, logFile, wsHost, wsPath)
}

func wsURLFromBase(baseURL string) (host, path string) {
	u, err := url.Parse(strings.TrimSuffix(baseURL, "/"))
	if err != nil {
		return "demo-api.kalshi.co", "/trade-api/ws/v2"
	}
	path = strings.Replace(u.Path, "/v2", "/ws/v2", 1)
	if path == u.Path {
		path = "/trade-api/ws/v2"
	}
	return u.Host, path
}

func logSection(log *os.File, title string, body string) {
	fmt.Fprintf(log, "=== %s ===\n", title)
	fmt.Fprintf(log, "%s\n\n", body)
}

func logJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func runAll(ctx context.Context, client *oddrip.Client, log *os.File, wsHost, wsPath string) {
	fullURL := fmt.Sprintf("wss://%s%s", wsHost, wsPath)
	logSection(log, "Connect", fmt.Sprintf("Full URL: %s\n(Connection established with auth headers from KALSHI_ACCESS_KEY / KALSHI_PRIVATE_KEY_PATH)", fullURL))

	conn, err := client.ConnectWS(ctx, oddrip.WSHost(wsHost), oddrip.WSPath(wsPath))
	if err != nil {
		logSection(log, "Connect (error)", err.Error())
		return
	}
	defer conn.Close()

	trueVal := true

	logCall := func(desc string, sent interface{}, fn func() (interface{}, error)) {
		fmt.Fprintf(log, "=== %s ===\n", desc)
		fmt.Fprintf(log, "Sent:\n%s\n", logJSON(sent))
		res, err := fn()
		if err != nil {
			fmt.Fprintf(log, "Response (error): %v\n\n", err)
			return
		}
		fmt.Fprintf(log, "Response:\n%s\n\n", logJSON(res))
	}

	var tickerSID int
	var orderbookSID int
	var subs []types.SubscribedResponse

	logCall("Subscribe (ticker, single market)",
		types.SubscribeCommand{Cmd: "subscribe", Params: types.SubscribeParams{
			Channels:     []string{types.WSChannelTicker},
			MarketTicker: exampleMarketTicker,
		}},
		func() (interface{}, error) {
			subs, err = conn.Subscribe(ctx, types.SubscribeParams{
				Channels:     []string{types.WSChannelTicker},
				MarketTicker: exampleMarketTicker,
			})
			if err != nil {
				return nil, err
			}
			if len(subs) > 0 {
				tickerSID = subs[0].Msg.SID
			}
			return subs, nil
		})

	logCall("Subscribe (orderbook_delta, single market)",
		types.SubscribeCommand{Cmd: "subscribe", Params: types.SubscribeParams{
			Channels:     []string{types.WSChannelOrderbookDelta},
			MarketTicker: exampleMarketTicker,
		}},
		func() (interface{}, error) {
			s, err := conn.Subscribe(ctx, types.SubscribeParams{
				Channels:     []string{types.WSChannelOrderbookDelta},
				MarketTicker: exampleMarketTicker,
			})
			if err != nil {
				return nil, err
			}
			if len(s) > 0 {
				orderbookSID = s[0].Msg.SID
			}
			return s, nil
		})

	logCall("Subscribe (trade, optional market)",
		types.SubscribeCommand{Cmd: "subscribe", Params: types.SubscribeParams{
			Channels:     []string{types.WSChannelTrade},
			MarketTicker: exampleMarketTicker,
		}},
		func() (interface{}, error) {
			return conn.Subscribe(ctx, types.SubscribeParams{
				Channels:     []string{types.WSChannelTrade},
				MarketTicker: exampleMarketTicker,
			})
		})

	logCall("Subscribe (market_lifecycle_v2, no market filter)",
		types.SubscribeCommand{Cmd: "subscribe", Params: types.SubscribeParams{
			Channels: []string{types.WSChannelMarketLifecycle},
		}},
		func() (interface{}, error) {
			return conn.Subscribe(ctx, types.SubscribeParams{
				Channels: []string{types.WSChannelMarketLifecycle},
			})
		})

	logCall("Subscribe (ticker, multiple markets, send_initial_snapshot)",
		types.SubscribeCommand{Cmd: "subscribe", Params: types.SubscribeParams{
			Channels:            []string{types.WSChannelTicker},
			MarketTickers:       []string{exampleMarketTicker, "KXETH-24DEC31-T5000"},
			SendInitialSnapshot: &trueVal,
		}},
		func() (interface{}, error) {
			return conn.Subscribe(ctx, types.SubscribeParams{
				Channels:            []string{types.WSChannelTicker},
				MarketTickers:       []string{exampleMarketTicker, "KXETH-24DEC31-T5000"},
				SendInitialSnapshot: &trueVal,
			})
		})

	logCall("ListSubscriptions",
		types.ListSubscriptionsCommand{Cmd: "list_subscriptions"},
		func() (interface{}, error) {
			return conn.ListSubscriptions(ctx)
		})

	listResp, _ := conn.ListSubscriptions(ctx)
	var updateSID int
	if listResp != nil && len(listResp.Msg) > 0 {
		updateSID = listResp.Msg[0].SID
		sidPtr := updateSID
		logCall("UpdateSubscription (add_markets)",
			types.UpdateSubscriptionCommand{Cmd: "update_subscription", Params: types.UpdateSubscriptionParams{
				SID:           &sidPtr,
				Action:        "add_markets",
				MarketTickers: []string{"KXETH-24DEC31-T4000"},
			}},
			func() (interface{}, error) {
				return conn.UpdateSubscription(ctx, types.UpdateSubscriptionParams{
					SID:           &sidPtr,
					Action:        "add_markets",
					MarketTickers: []string{"KXETH-24DEC31-T4000"},
				})
			})
	}

	fmt.Fprintf(log, "=== Receive messages (5s) ===\n")
	deadline := time.After(5 * time.Second)
	received := 0
loop:
	for {
		select {
		case <-deadline:
			break loop
		case msg, ok := <-conn.Messages():
			if !ok {
				break loop
			}
			received++
			fmt.Fprintf(log, "Message #%d type=%s sid=%d seq=%d\n", received, msg.Type, msg.SID, msg.Seq)
			fmt.Fprintf(log, "Raw msg:\n%s\n\n", string(msg.Msg))
		}
	}
	fmt.Fprintf(log, "Received %d messages.\n\n", received)

	if updateSID != 0 {
		sidPtr := updateSID
		logCall("UpdateSubscription (delete_markets)",
			types.UpdateSubscriptionCommand{Cmd: "update_subscription", Params: types.UpdateSubscriptionParams{
				SID:           &sidPtr,
				Action:        "delete_markets",
				MarketTickers: []string{"KXETH-24DEC31-T4000"},
			}},
			func() (interface{}, error) {
				return conn.UpdateSubscription(ctx, types.UpdateSubscriptionParams{
					SID:           &sidPtr,
					Action:        "delete_markets",
					MarketTickers: []string{"KXETH-24DEC31-T4000"},
				})
			})
	}

	if tickerSID != 0 {
		cmd := types.UnsubscribeCommand{Cmd: "unsubscribe"}
		cmd.Params.Sids = []int{tickerSID}
		logCall("Unsubscribe (ticker sid)", cmd, func() (interface{}, error) {
			return nil, conn.Unsubscribe(ctx, []int{tickerSID})
		})
	}
	if orderbookSID != 0 {
		cmd := types.UnsubscribeCommand{Cmd: "unsubscribe"}
		cmd.Params.Sids = []int{orderbookSID}
		logCall("Unsubscribe (orderbook_delta sid)", cmd, func() (interface{}, error) {
			return nil, conn.Unsubscribe(ctx, []int{orderbookSID})
		})
	}

	logSection(log, "Close", "Connection closed.")
}
