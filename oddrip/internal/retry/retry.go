package retry

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/UTXOnly/oddrip/oddrip/internal/errors"
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

func Do(ctx context.Context, cfg Config, fn func() (*http.Response, error)) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		resp, err := fn()
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if attempt < cfg.MaxAttempts-1 {
				time.Sleep(cfg.Delay(attempt, 0))
			}
			continue
		}
		if resp.StatusCode < 400 || !errors.IsRetryable(resp.StatusCode) {
			return resp, nil
		}
		var retryAfter time.Duration
		if s := resp.Header.Get("Retry-After"); s != "" {
			if sec, err := strconv.Atoi(s); err == nil {
				retryAfter = time.Duration(sec) * time.Second
			}
		}
		resp.Body.Close()
		if attempt == cfg.MaxAttempts-1 {
			return nil, lastErr
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		time.Sleep(cfg.Delay(attempt, retryAfter))
	}
	return nil, lastErr
}
