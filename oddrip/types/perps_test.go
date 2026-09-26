package types

import (
	"encoding/json"
	"testing"
)

func TestTickerPrice_JSON(t *testing.T) {
	const payload = `{"price":"65000.1234","ts_ms":1700000000123}`
	var p TickerPrice
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		t.Fatal(err)
	}
	if p.Price != "65000.1234" || p.TsMs != 1700000000123 {
		t.Fatalf("unexpected: %+v", p)
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != payload {
		t.Fatalf("marshal: got %s, want %s", b, payload)
	}
}
