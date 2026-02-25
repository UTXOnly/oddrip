package errors

import (
	"fmt"
	"io"
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

func ParseResponseError(statusCode int, body io.Reader, requestID string) *APIError {
	err := &APIError{StatusCode: statusCode, RequestID: requestID}
	if body == nil {
		return err
	}
	buf, _ := io.ReadAll(io.LimitReader(body, maxBodySnippet))
	if len(buf) > 0 {
		err.RawBody = string(buf)
	}
	return err
}

func ParseJSONError(statusCode int, code, message, details, service, requestID string, rawBody string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
		Service:    service,
		RequestID:  requestID,
		RawBody:    rawBody,
	}
}

func IsRetryable(statusCode int) bool {
	return statusCode == 429 || (statusCode >= 500 && statusCode < 600)
}
