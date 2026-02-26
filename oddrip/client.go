package oddrip

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/UTXOnly/oddrip/oddrip/internal/errors"
	"github.com/UTXOnly/oddrip/oddrip/internal/retry"
	"github.com/UTXOnly/oddrip/oddrip/types"
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

	Exchange  *ExchangeService
	Markets   *MarketsService
	Orders    *OrdersService
	Portfolio *PortfolioService
	Account   *AccountService
	Events    *EventsService
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
				IdleConnTimeout:     90e9,
			},
			Timeout: 30e9,
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
	return c
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body interface{}, out interface{}) error {
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

	var resp *http.Response
	resp, doErr := retry.Do(ctx, c.retry, func() (*http.Response, error) {
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
	if doErr != nil {
		return doErr
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get("Request-Id")
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := errors.ParseResponseError(resp.StatusCode, resp.Body, requestID)
		dec := json.NewDecoder(strings.NewReader(apiErr.RawBody))
		var er types.ErrorResponse
		if dec.Decode(&er) == nil {
			apiErr.Code = er.Code
			apiErr.Message = er.Message
			apiErr.Details = er.Details
			apiErr.Service = er.Service
		}
		return wrapAPIError(apiErr)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out interface{}) error {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodPost, path, nil, body, out)
}

func (c *Client) put(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodPut, path, nil, body, out)
}

func (c *Client) delete(ctx context.Context, path string, query url.Values, body interface{}, out interface{}) error {
	return c.do(ctx, http.MethodDelete, path, query, body, out)
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

func joinPath(elem ...string) string {
	return "/" + path.Join(elem...)
}
