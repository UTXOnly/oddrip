package types

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Dollars is a fixed-point dollar amount scaled by 1e6: one unit is
// $0.000001. The API's FixedPointDollars strings carry up to six decimals in
// responses (fees, fill costs) and two to four in requests (prices), so this
// scale is lossless for every value the exchange emits.
type Dollars int64

// Count is a fixed-point contract count scaled by 100: one unit is 0.01
// contracts, the minimum granularity of the API's FixedPointCount strings.
type Count int64

const (
	dollarsFrac  = 6
	dollarsScale = 1_000_000
	countFrac    = 2
	countScale   = 100
)

// ParseDollars parses a FixedPointDollars string such as "0.4500", "12",
// "-3.25", or "0.010000". At most six decimal places are accepted.
func ParseDollars(s string) (Dollars, error) {
	n, err := parseFixed(s, dollarsFrac)
	if err != nil {
		return 0, fmt.Errorf("parse dollars %q: %w", s, err)
	}
	return Dollars(n), nil
}

// String formats with four decimals ("0.4500"), the form request fields
// accept. Values with non-zero digits in the fifth or sixth place, such as
// fees, are formatted with six.
func (d Dollars) String() string {
	if d%100 == 0 {
		return formatFixed(int64(d)/100, 4)
	}
	return formatFixed(int64(d), dollarsFrac)
}

func (d Dollars) Float64() float64 { return float64(d) / dollarsScale }

// Cents truncates toward zero: $0.4567 is 45 cents, -$0.4567 is -45.
func (d Dollars) Cents() int64 { return int64(d) / (dollarsScale / 100) }

// ParseCount parses a FixedPointCount string such as "10", "10.5", or
// "136.00". At most two decimal places are accepted.
func ParseCount(s string) (Count, error) {
	n, err := parseFixed(s, countFrac)
	if err != nil {
		return 0, fmt.Errorf("parse count %q: %w", s, err)
	}
	return Count(n), nil
}

// String formats with two decimals ("10.00"), matching API responses.
func (c Count) String() string { return formatFixed(int64(c), countFrac) }

func (c Count) Float64() float64 { return float64(c) / countScale }

func parseFixed(s string, maxFrac int) (int64, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	digits := s
	neg := digits[0] == '-'
	if neg {
		digits = digits[1:]
	}
	intPart, fracPart, hasDot := strings.Cut(digits, ".")
	if intPart == "" || (hasDot && fracPart == "") {
		return 0, errors.New("malformed number")
	}
	if len(fracPart) > maxFrac {
		return 0, fmt.Errorf("more than %d decimal places", maxFrac)
	}
	var n int64
	for _, part := range [2]string{intPart, fracPart + strings.Repeat("0", maxFrac-len(fracPart))} {
		for i := 0; i < len(part); i++ {
			c := part[i]
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("invalid character %q", c)
			}
			d := int64(c - '0')
			if n > (math.MaxInt64-d)/10 {
				return 0, errors.New("overflow")
			}
			n = n*10 + d
		}
	}
	if neg {
		n = -n
	}
	return n, nil
}

func formatFixed(n int64, frac int) string {
	u := uint64(n)
	sign := ""
	if n < 0 {
		u, sign = -u, "-"
	}
	pow := uint64(1)
	for i := 0; i < frac; i++ {
		pow *= 10
	}
	return fmt.Sprintf("%s%d.%0*d", sign, u/pow, frac, u%pow)
}

// ParseTime parses the RFC3339 timestamps the API emits: with or without
// fractional seconds, with a Z or numeric offset. A timestamp with no zone
// designator is treated as UTC.
func ParseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}
	if t, err2 := time.Parse("2006-01-02T15:04:05.999999999", s); err2 == nil {
		return t, nil
	}
	return time.Time{}, err
}
