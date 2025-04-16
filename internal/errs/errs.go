package errs

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kymppi/kuura/internal/errcode"
)

type Error struct {
	Code     errcode.ErrorCode
	Err      error
	TraceID  string
	Metadata map[string]json.RawMessage
}

func New(code errcode.ErrorCode, err error) *Error {
	return &Error{
		Code: code,
		Err:  err,
	}
}

func (e *Error) WithMetadata(key string, value any) *Error {
	if e.Metadata == nil {
		e.Metadata = make(map[string]json.RawMessage)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return e
	}

	e.Metadata[key] = data

	return e
}

func (e *Error) WithTraceID(traceID string) *Error {
	e.TraceID = traceID
	return e
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s (trace: %s)", e.Code, e.Err.Error(), e.TraceID)
	}
	return fmt.Sprintf("[%s] Unknown error (trace: %s)", e.Code, e.TraceID)
}

// support errors.Is and errors.As
func (e *Error) Unwrap() error {
	return e.Err
}

func IsErrorCode(err error, code errcode.ErrorCode) bool {
	var customErr *Error
	return errors.As(err, &customErr) && customErr.Code == code
}
