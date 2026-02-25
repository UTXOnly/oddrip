package oddrip

import (
	"fmt"

	"github.com/oddrip/client/oddrip/internal/errors"
)

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

func wrapAPIError(e *errors.APIError) *APIError {
	if e == nil {
		return nil
	}
	return &APIError{
		StatusCode: e.StatusCode,
		Code:       e.Code,
		Message:    e.Message,
		Details:    e.Details,
		Service:    e.Service,
		RequestID:  e.RequestID,
		RawBody:    e.RawBody,
	}
}
