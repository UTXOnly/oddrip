package types

import "encoding/json"

// EventLiveData is the live data payload attached to an event. Type names the
// schema of Details (crypto price charts, commodity timeseries, weather
// observations, etc.), which is left raw because the shape varies by type.
type EventLiveData struct {
	Type         string          `json:"type"`
	Details      json.RawMessage `json:"details"`
	IsHistorical *bool           `json:"is_historical,omitempty"`
	DefaultRange string          `json:"default_range,omitempty"`
	RangeOptions []string        `json:"range_options,omitempty"`
}

type GetEventLiveDataResponse struct {
	LiveData EventLiveData `json:"live_data"`
}

type GetEventLiveDataOpts struct {
	// Range is a chart range hint such as "15min", "1h", or "1d".
	Range string
}

// WeatherIndexStationReading is one member station's reported reading and
// quality-control disposition for a minute of the index, returned only when
// GetWeatherIndexOpts.Detailed is set.
type WeatherIndexStationReading struct {
	StationID    string   `json:"station_id"`
	Code         string   `json:"code"`
	Source       string   `json:"source,omitempty"`
	TempF        *float64 `json:"temp_f,omitempty"`
	ObsTimeMs    *int64   `json:"obs_time_ms,omitempty"`
	ReceivedAtMs *int64   `json:"received_at_ms,omitempty"`
	PrimaryCode  string   `json:"primary_code,omitempty"`
}

// WeatherIndexPoint is one minute of the city index. V is absent for minutes
// where the index quorum failed; such minutes are omitted from the series, so
// gaps are real gaps.
type WeatherIndexPoint struct {
	T            int64                        `json:"t"`
	V            *float64                     `json:"v,omitempty"`
	Status       string                       `json:"status"`
	Contributors *int                         `json:"contributors,omitempty"`
	ReceiptBasis string                       `json:"receipt_basis,omitempty"`
	Stations     []WeatherIndexStationReading `json:"stations,omitempty"`
}

type GetWeatherIndexResponse struct {
	City          string              `json:"city"`
	ConfigVersion string              `json:"config_version,omitempty"`
	Units         string              `json:"units"`
	Timeseries    []WeatherIndexPoint `json:"timeseries"`
}

// GetWeatherIndexOpts bounds the returned window. From/To are unix
// milliseconds and are inclusive; LastSec is a trailing window in seconds and
// is mutually exclusive with From/To.
type GetWeatherIndexOpts struct {
	From     *int64
	To       *int64
	LastSec  *int64
	Detailed *bool
}

type WeatherIndexCalibrationStation struct {
	StationID  string  `json:"station_id"`
	Weight     float64 `json:"weight"`
	OffsetC    float64 `json:"offset_c"`
	UpdateNote string  `json:"update_note,omitempty"`
}

// WeatherIndexCalibration is one version of a city's index configuration: the
// weights, offsets, and city reference used from EffectiveAtMs until the next
// record.
type WeatherIndexCalibration struct {
	ConfigVersion            string                           `json:"config_version"`
	PublishedAtMs            *int64                           `json:"published_at_ms,omitempty"`
	EffectiveAtMs            int64                            `json:"effective_at_ms"`
	ChangeReason             string                           `json:"change_reason,omitempty"`
	CalibrationWindowStartMs *int64                           `json:"calibration_window_start_ms,omitempty"`
	CalibrationWindowEndMs   *int64                           `json:"calibration_window_end_ms,omitempty"`
	CityReferenceC           float64                          `json:"city_reference_c"`
	Stations                 []WeatherIndexCalibrationStation `json:"stations"`
}

type GetWeatherIndexCalibrationsResponse struct {
	City         string                    `json:"city"`
	Units        string                    `json:"units"`
	Calibrations []WeatherIndexCalibration `json:"calibrations"`
}
