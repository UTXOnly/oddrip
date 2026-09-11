package types

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestParseDollars(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Dollars
		str  string
	}{
		{"0.5600", 560000, "0.5600"},
		{"0.56", 560000, "0.5600"},
		{"0.4500", 450000, "0.4500"},
		{"0.480", 480000, "0.4800"},
		{"0.35", 350000, "0.3500"},
		{"50.0000", 50000000, "50.0000"},
		{"100.0000", 100000000, "100.0000"},
		{"0", 0, "0.0000"},
		{"0.5", 500000, "0.5000"},
		{"1", 1000000, "1.0000"},
		{"99.9999", 99999900, "99.9999"},
		{"-3.25", -3250000, "-3.2500"},
		{"-0.0001", -100, "-0.0001"},
		{"0.010000", 10000, "0.0100"},
		{"0.000000", 0, "0.0000"},
		{"0.000001", 1, "0.000001"},
		{"0.123456", 123456, "0.123456"},
		{"-0.123456", -123456, "-0.123456"},
		{"1234567.891011", 1234567891011, "1234567.891011"},
	} {
		got, err := ParseDollars(tc.in)
		if err != nil {
			t.Errorf("%q: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%q: got %d want %d", tc.in, got, tc.want)
		}
		if s := got.String(); s != tc.str {
			t.Errorf("%q: String() = %q want %q", tc.in, s, tc.str)
		}
		back, err := ParseDollars(got.String())
		if err != nil || back != got {
			t.Errorf("%q: round trip %q -> %d, %v", tc.in, got.String(), back, err)
		}
	}
}

func TestParseDollars_Reject(t *testing.T) {
	for _, in := range []string{
		"", "-", ".", "-.", "+1", "+1.00", "--1", "1.", ".5", "1..0", "1.2.3",
		"abc", "1a", "a1", "0x10", "1e5", " 1", "1 ", "$1.00", "1,000.00",
		"0.1234567", "1.0000000",
		"9223372036854775808", "99999999999999999999",
	} {
		if got, err := ParseDollars(in); err == nil {
			t.Errorf("%q: want error, got %d", in, got)
		}
	}
	_, err := ParseDollars("0.1234567")
	if err == nil || !strings.Contains(err.Error(), "6 decimal places") {
		t.Errorf("precision error: %v", err)
	}
	_, err = ParseDollars("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("empty error: %v", err)
	}
}

func TestDollars_Cents(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int64
	}{
		{"0.4500", 45},
		{"0.4567", 45},
		{"0.4599", 45},
		{"0.009999", 0},
		{"0.01", 1},
		{"1", 100},
		{"99.9999", 9999},
		{"-3.25", -325},
		{"-0.4567", -45},
		{"-0.009999", 0},
	} {
		d, err := ParseDollars(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if got := d.Cents(); got != tc.want {
			t.Errorf("%q: Cents() = %d want %d", tc.in, got, tc.want)
		}
	}
}

func TestDollars_Float64(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want float64
	}{
		{"0.4500", 0.45},
		{"1", 1},
		{"-3.25", -3.25},
		{"0.000001", 0.000001},
	} {
		d, err := ParseDollars(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if got := d.Float64(); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%q: Float64() = %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseCount(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Count
		str  string
	}{
		{"10.00", 1000, "10.00"},
		{"10.0", 1000, "10.00"},
		{"10", 1000, "10.00"},
		{"2.50", 250, "2.50"},
		{"0.01", 1, "0.01"},
		{"0", 0, "0.00"},
		{"0.5", 50, "0.50"},
		{"1", 100, "1.00"},
		{"136.00", 13600, "136.00"},
		{"33896.00", 3389600, "33896.00"},
		{"-54.00", -5400, "-54.00"},
		{"-0.01", -1, "-0.01"},
	} {
		got, err := ParseCount(tc.in)
		if err != nil {
			t.Errorf("%q: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%q: got %d want %d", tc.in, got, tc.want)
		}
		if s := got.String(); s != tc.str {
			t.Errorf("%q: String() = %q want %q", tc.in, s, tc.str)
		}
		back, err := ParseCount(got.String())
		if err != nil || back != got {
			t.Errorf("%q: round trip %q -> %d, %v", tc.in, got.String(), back, err)
		}
	}
}

func TestParseCount_Reject(t *testing.T) {
	for _, in := range []string{
		"", "-", "+10", "10.", ".5", "10.000", "0.001", "1..0", "ten", "1e2", " 10", "10 ",
	} {
		if got, err := ParseCount(in); err == nil {
			t.Errorf("%q: want error, got %d", in, got)
		}
	}
	_, err := ParseCount("10.000")
	if err == nil || !strings.Contains(err.Error(), "2 decimal places") {
		t.Errorf("precision error: %v", err)
	}
}

func TestCount_Float64(t *testing.T) {
	c, err := ParseCount("2.50")
	if err != nil {
		t.Fatal(err)
	}
	if c.Float64() != 2.5 {
		t.Errorf("Float64() = %v", c.Float64())
	}
}

func TestParseTime(t *testing.T) {
	utc := time.Date(2022, 11, 22, 20, 44, 1, 0, time.UTC)
	for _, tc := range []struct {
		in   string
		want time.Time
	}{
		{"2022-11-22T20:44:01Z", utc},
		{"2022-11-22T20:44:01.5Z", utc.Add(500 * time.Millisecond)},
		{"2022-11-22T20:44:01.123456Z", utc.Add(123456 * time.Microsecond)},
		{"2022-11-22T20:44:01.123456789Z", utc.Add(123456789)},
		{"2022-11-22T15:44:01-05:00", utc},
		{"2022-11-22T21:44:01.25+01:00", utc.Add(250 * time.Millisecond)},
		{"2022-11-22T20:44:01+00:00", utc},
		{"2022-11-22T20:44:01", utc},
		{"2022-11-22T20:44:01.123456", utc.Add(123456 * time.Microsecond)},
	} {
		got, err := ParseTime(tc.in)
		if err != nil {
			t.Errorf("%q: %v", tc.in, err)
			continue
		}
		if !got.Equal(tc.want) {
			t.Errorf("%q: got %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseTime_Reject(t *testing.T) {
	for _, in := range []string{
		"", "garbage", "2022-11-22", "20:44:01", "2022-11-22 20:44:01Z",
		"1669149841", "2022-13-01T00:00:00Z", "2022-11-22T25:00:00Z",
	} {
		if got, err := ParseTime(in); err == nil {
			t.Errorf("%q: want error, got %v", in, got)
		}
	}
	_, err := ParseTime("garbage")
	if err == nil || !strings.Contains(err.Error(), "garbage") {
		t.Errorf("error should quote input: %v", err)
	}
}
