package errors

import "fmt"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func WithMessage(err *Error, message string) *Error {
	return &Error{Code: err.Code, Message: message}
}

var (
	ErrInvalidParams  = New(40000, "invalid parameters")
	ErrUnauthorized   = New(40100, "unauthorized")
	ErrForbidden      = New(40300, "forbidden")
	ErrNotFound       = New(40400, "not found")
	ErrInternal       = New(50000, "internal error")
	ErrPipelineNotFound = New(40401, "pipeline not found")
	ErrPromptNotFound   = New(40402, "prompt template not found")
	ErrJobNotFound      = New(40403, "job not found")
	ErrProviderError    = New(50001, "llm provider error")
)
