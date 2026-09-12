package oddrip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/UTXOnly/oddrip/oddrip/types"
)

const (
	// maxBodySnippet bounds APIError.RawBody.
	maxBodySnippet = 512
	// maxErrorBody bounds how much of an error body is read to decode the
	// structured fields. A body longer than this is truncated, so it does not
	// parse as JSON and the fields stay empty.
	maxErrorBody = 64 << 10
)

// APIError is a non-2xx response. StatusCode and RawBody (the first 512 bytes
// of the body) are always set. Code, Message, and Details are filled from the
// body when it is one of the shapes Kalshi emits: the spec's flat
// ErrorResponse, the same object nested under "error" (what production
// returns for most errors), or {"msg": "..."} (parameter-binding 400s). They
// are decoded from up to 64 KiB of the body; bodies over 64 KiB are not
// decoded and leave them empty. RequestID is the Request-Id header when
// present; production does not currently send one.
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

// errorBody covers every error body shape observed from the API. The flat
// fields are the spec's ErrorResponse; Error is the production wrapper; Msg is
// the parameter-binding validator's shape.
type errorBody struct {
	types.ErrorResponse
	Error *types.ErrorResponse `json:"error,omitempty"`
	Msg   string               `json:"msg,omitempty"`
}

func newAPIError(resp *http.Response) *APIError {
	e := &APIError{StatusCode: resp.StatusCode, RequestID: resp.Header.Get("Request-Id")}
	buf, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	e.RawBody = string(buf[:min(len(buf), maxBodySnippet)])
	var body errorBody
	if json.NewDecoder(bytes.NewReader(buf)).Decode(&body) != nil {
		return e
	}
	er := body.ErrorResponse
	if er.Code == "" && er.Message == "" && body.Error != nil {
		er = *body.Error
	}
	e.Code = er.Code
	e.Message = er.Message
	e.Details = er.Details
	e.Service = er.Service
	if e.Message == "" && body.Msg != "" {
		e.Message = body.Msg
	}
	return e
}
