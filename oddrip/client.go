package oddrip

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UTXOnly/oddrip/oddrip/internal/retry"
)

const defaultBaseURL = "https://api.elections.kalshi.com/trade-api/v2"

type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	JitterPercent float64
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       AuthProvider
	retry      retry.Config

	Exchange    *ExchangeService
	Markets     *MarketsService
	Orders      *OrdersService
	Portfolio   *PortfolioService
	Account     *AccountService
	Events      *EventsService
	LiveData    *LiveDataService
	Series      *SeriesService
	OrderGroups *OrderGroupsService
	Subaccounts *SubaccountsService
}

type Option func(*Client)

func BaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(u, "/")
	}
}

func HTTPClient(h *http.Client) Option {
	return func(c *Client) {
		c.httpClient = h
	}
}

func Auth(p AuthProvider) Option {
	return func(c *Client) {
		c.auth = p
	}
}

func RetryConfigOption(cfg RetryConfig) Option {
	return func(c *Client) {
		c.retry = retry.Config{
			MaxAttempts:   cfg.MaxAttempts,
			InitialDelay:  cfg.InitialDelay,
			MaxDelay:      cfg.MaxDelay,
			JitterPercent: cfg.JitterPercent,
		}
	}
}

func New(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: 30 * time.Second,
		},
		retry: retry.DefaultConfig,
	}
	for _, o := range opts {
		o(c)
	}
	c.Exchange = &ExchangeService{client: c}
	c.Markets = &MarketsService{client: c}
	c.Orders = &OrdersService{client: c}
	c.Portfolio = &PortfolioService{client: c}
	c.Account = &AccountService{client: c}
	c.Events = &EventsService{client: c}
	c.LiveData = &LiveDataService{client: c}
	c.Series = &SeriesService{client: c}
	c.OrderGroups = &OrderGroupsService{client: c}
	c.Subaccounts = &SubaccountsService{client: c}
	return c
}

// ErrEmptyPathParam is returned when a path parameter (ticker, order ID,
// series ticker, ...) is empty or a "." / ".." segment. Dropping the segment
// would route the request to a different endpoint — an empty order ID would
// turn DELETE /portfolio/events/orders/{id} into CancelAll — so the request is
// refused before anything is sent.
var ErrEmptyPathParam = errors.New("oddrip: empty path parameter")

// do sends one request. idempotent selects the retry policy: idempotent
// requests are retried on 429, 5xx, and transport errors; non-idempotent ones
// only on 429, because a 5xx or a dropped connection may mean the write was
// applied and a replay would apply it again. GET, PUT, and DELETE are
// idempotent by construction; POST is only when the body carries a
// deduplication key (see postIdempotent).
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body interface{}, out interface{}, idempotent bool) error {
	if path == "" {
		return ErrEmptyPathParam
	}
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	resp, err := retry.Do(ctx, c.retry, idempotent, func() (*http.Response, error) {
		var bodyReader io.Reader
		if len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
		if err != nil {
			return nil, err
		}
		if len(bodyBytes) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")
		if c.auth != nil {
			if err := c.auth.Apply(req); err != nil {
				return nil, err
			}
		}
		return c.httpClient.Do(req)
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newAPIError(resp)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out interface{}) error {
	return c.do(ctx, http.MethodGet, path, query, nil, out, true)
}

// post is for writes that are not safe to replay (amend, decrease, create
// without a deduplication key): retried on 429 only.
func (c *Client) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.postQuery(ctx, path, nil, body, out)
}

func (c *Client) postQuery(ctx context.Context, path string, query url.Values, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodPost, path, query, body, out, false)
}

// postIdempotent is for POSTs the server deduplicates (client_order_id,
// client_transfer_id) or that set absolute state: full retry policy.
func (c *Client) postIdempotent(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodPost, path, nil, body, out, true)
}

func (c *Client) put(ctx context.Context, path string, query url.Values, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodPut, path, query, body, out, true)
}

func (c *Client) delete(ctx context.Context, path string, query url.Values, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodDelete, path, query, body, out, true)
}

func encodeQuery(v url.Values, key string, value string) {
	if value != "" {
		v.Set(key, value)
	}
}

func encodeQueryInt64(v url.Values, key string, p *int64) {
	if p != nil {
		v.Set(key, fmt.Sprintf("%d", *p))
	}
}

func encodeQueryInt(v url.Values, key string, p *int) {
	if p != nil {
		v.Set(key, fmt.Sprintf("%d", *p))
	}
}

func encodeQueryBool(v url.Values, key string, p *bool) {
	if p != nil {
		v.Set(key, fmt.Sprintf("%t", *p))
	}
}

func encodeQueryStrings(v url.Values, key string, values []string) {
	for _, s := range values {
		if s != "" {
			v.Add(key, s)
		}
	}
}

// joinPath builds a request path from literal and parameter segments. Each
// segment is path-escaped so a value containing "/" stays a single segment.
// It returns "" when any segment is empty, ".", or "..", and do() rejects that
// with ErrEmptyPathParam.
func joinPath(elem ...string) string {
	var b strings.Builder
	for _, e := range elem {
		if e == "" || e == "." || e == ".." {
			return ""
		}
		b.WriteByte('/')
		b.WriteString(url.PathEscape(e))
	}
	return b.String()
}
