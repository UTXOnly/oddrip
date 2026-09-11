package oddrip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

const maxBodySnippet = 512

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    string
	Service    string
	RequestID  string
	RawBody    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("api error %d", e.StatusCode)
}

func newAPIError(resp *http.Response) *APIError {
	e := &APIError{StatusCode: resp.StatusCode, RequestID: resp.Header.Get("Request-Id")}
	buf, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodySnippet))
	e.RawBody = string(buf)
	var er types.ErrorResponse
	if json.NewDecoder(bytes.NewReader(buf)).Decode(&er) == nil {
		e.Code = er.Code
		e.Message = er.Message
		e.Details = er.Details
		e.Service = er.Service
	}
	return e
}
