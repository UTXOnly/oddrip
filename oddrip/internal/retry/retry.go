package retry

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	JitterPercent float64
}

var DefaultConfig = Config{
	MaxAttempts:   4,
	InitialDelay:  500 * time.Millisecond,
	MaxDelay:      30 * time.Second,
	JitterPercent: 0.2,
}

func (c Config) Delay(attempt int, retryAfter time.Duration) time.Duration {
	var d time.Duration
	if retryAfter > 0 {
		d = retryAfter
		if d > c.MaxDelay {
			d = c.MaxDelay
		}
	} else {
		backoff := c.InitialDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > c.MaxDelay {
			backoff = c.MaxDelay
		}
		jitter := float64(backoff) * c.JitterPercent * (2*rand.Float64() - 1)
		d = backoff + time.Duration(jitter)
		if d < 0 {
			d = c.InitialDelay
		}
	}
	return d
}

func IsRetryable(statusCode int) bool {
	return statusCode == 429 || (statusCode >= 500 && statusCode < 600)
}

// Do never returns (nil, nil). When every attempt yields a retryable status
// the last response is returned with its body open so the caller can parse it.
func Do(ctx context.Context, cfg Config, fn func() (*http.Response, error)) (*http.Response, error) {
	attempts := cfg.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 0; ; attempt++ {
		resp, err := fn()
		last := attempt == attempts-1
		var retryAfter time.Duration
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if last {
				return nil, err
			}
		} else {
			if last || !IsRetryable(resp.StatusCode) {
				return resp, nil
			}
			retryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
		}
		if err := wait(ctx, cfg.Delay(attempt, retryAfter)); err != nil {
			return nil, err
		}
	}
}

func parseRetryAfter(s string) time.Duration {
	if s == "" {
		return 0
	}
	if sec, err := strconv.Atoi(s); err == nil {
		return time.Duration(sec) * time.Second
	}
	if t, err := http.ParseTime(s); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

func wait(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
