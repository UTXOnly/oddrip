package types

import (
	"encoding/json"
	"testing"
)

func TestGetWeatherIndexResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"city": "miami",
		"config_version": "2026-08-31",
		"units": "F",
		"timeseries": [
			{"t": 1716300000000, "v": 81.25, "status": "ok", "contributors": 3, "receipt_basis": "obs_time"},
			{"t": 1716300060000, "status": "no_quorum"}
		]
	}`
	var out GetWeatherIndexResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.City != "miami" || out.Units != "F" || out.ConfigVersion != "2026-08-31" {
		t.Fatalf("unexpected: %+v", out)
	}
	if len(out.Timeseries) != 2 {
		t.Fatalf("timeseries: %+v", out.Timeseries)
	}
	if out.Timeseries[0].V == nil || *out.Timeseries[0].V != 81.25 {
		t.Fatalf("first point: %+v", out.Timeseries[0])
	}
	if out.Timeseries[1].V != nil {
		t.Fatalf("quorum-failure point should carry no value: %+v", out.Timeseries[1])
	}
}

func TestWeatherIndexPoint_DetailedStations(t *testing.T) {
	const payload = `{
		"t": 1716300000000,
		"v": 81.25,
		"status": "ok",
		"stations": [
			{"station_id": "KMIA", "code": "accepted", "source": "metar", "temp_f": 81.5, "obs_time_ms": 1716299970000, "received_at_ms": 1716300005000, "primary_code": "ok"},
			{"station_id": "KOPF", "code": "rejected_stale"}
		]
	}`
	var out WeatherIndexPoint
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Stations) != 2 || out.Stations[0].TempF == nil || *out.Stations[0].TempF != 81.5 {
		t.Fatalf("stations: %+v", out.Stations)
	}
	if out.Stations[1].Code != "rejected_stale" || out.Stations[1].TempF != nil {
		t.Fatalf("second station: %+v", out.Stations[1])
	}
}

func TestGetWeatherIndexCalibrationsResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"city": "miami",
		"units": "F",
		"calibrations": [{
			"config_version": "2026-08-31",
			"published_at_ms": 1716300000000,
			"effective_at_ms": 1716303600000,
			"change_reason": "weekly_offset_calibration",
			"calibration_window_start_ms": 1715694000000,
			"calibration_window_end_ms": 1716298800000,
			"city_reference_c": 27.5,
			"stations": [{"station_id": "KMIA", "weight": 0.6, "offset_c": -0.12, "update_note": "re-estimated"}]
		}]
	}`
	var out GetWeatherIndexCalibrationsResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Calibrations) != 1 {
		t.Fatalf("calibrations: %+v", out.Calibrations)
	}
	c := out.Calibrations[0]
	if c.CityReferenceC != 27.5 || c.EffectiveAtMs != 1716303600000 {
		t.Fatalf("calibration: %+v", c)
	}
	if len(c.Stations) != 1 || c.Stations[0].Weight != 0.6 || c.Stations[0].OffsetC != -0.12 {
		t.Fatalf("stations: %+v", c.Stations)
	}
}

func TestGetEventLiveDataResponse_Unmarshal(t *testing.T) {
	const payload = `{
		"live_data": {
			"type": "crypto_price_chart",
			"details": {"points": [[1716300000000, 65000.5]]},
			"is_historical": false,
			"default_range": "1h",
			"range_options": ["15min", "1h", "1d"]
		}
	}`
	var out GetEventLiveDataResponse
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if out.LiveData.Type != "crypto_price_chart" || len(out.LiveData.RangeOptions) != 3 {
		t.Fatalf("live_data: %+v", out.LiveData)
	}
	if string(out.LiveData.Details) != `{"points": [[1716300000000, 65000.5]]}` {
		t.Fatalf("details: %s", out.LiveData.Details)
	}
}
