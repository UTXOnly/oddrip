package types

import (
	"encoding/json"
	"testing"
)

func TestSubscribeCommand_Marshal(t *testing.T) {
	cmd := SubscribeCommand{
		ID:  1,
		Cmd: "subscribe",
		Params: SubscribeParams{
			Channels:     []string{WSChannelTicker, WSChannelOrderbookDelta},
			MarketTicker: "FED-23DEC-T3.00",
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if json.Unmarshal(data, &m) != nil {
		t.Fatal("unmarshal")
	}
	if m["cmd"] != "subscribe" || m["id"].(float64) != 1 {
		t.Errorf("unexpected cmd or id: %v", m)
	}
	params := m["params"].(map[string]interface{})
	channels := params["channels"].([]interface{})
	if len(channels) != 2 || params["market_ticker"] != "FED-23DEC-T3.00" {
		t.Errorf("params: %v", params)
	}
}

func TestUpdateSubscriptionParams_Action(t *testing.T) {
	cmd := UpdateSubscriptionCommand{
		ID:  2,
		Cmd: "update_subscription",
		Params: UpdateSubscriptionParams{
			SID:    intPtr(10),
			Action: "add_markets",
			MarketTickers: []string{"TICK-A", "TICK-B"},
		},
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	var decoded UpdateSubscriptionCommand
	if json.Unmarshal(data, &decoded) != nil {
		t.Fatal("unmarshal")
	}
	if decoded.Params.Action != "add_markets" || *decoded.Params.SID != 10 || len(decoded.Params.MarketTickers) != 2 {
		t.Errorf("decoded: %+v", decoded)
	}
}

func intPtr(i int) *int { return &i }
