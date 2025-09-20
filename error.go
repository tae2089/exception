package exception

import (
	"fmt"
	"runtime"
	"strings"
)

type ErrorCode int

const (
	ErrorBadRequest         ErrorCode = 400
	ErrorUnauthorized       ErrorCode = 401
	ErrorForbidden          ErrorCode = 403
	ErrorNotFound           ErrorCode = 404
	ErrorInternalServer     ErrorCode = 500
	ErrorNotImplemented     ErrorCode = 501
	ErrorServiceUnavailable ErrorCode = 503
)

type CustomError struct {
	code           ErrorCode
	Message        string   `json:"message"`
	Trace          string   `json:"trace"`
	PreviousTraces []string `json:"previous_traces"`
	Err            error    `json:"-"`
}

type CustomErrorOption func(*CustomError)

func WithCode(code ErrorCode) CustomErrorOption {
	return func(e *CustomError) { e.code = code }
}
func WithMessage(msg string) CustomErrorOption {
	return func(e *CustomError) { e.Message = msg }
}
func WithTrace(trace string) CustomErrorOption {
	return func(e *CustomError) { e.Trace = trace }
}
func WithPreviousTraces(traces []string) CustomErrorOption {
	return func(e *CustomError) { e.PreviousTraces = traces }
}
func WithCause(err error) CustomErrorOption {
	return func(e *CustomError) { e.Err = err }
}

func (e *CustomError) Error() string {
	if e.Message == "" {
		return "unknown error"
	}
	return e.Message
}

func (e *CustomError) Cause() error {
	return e.Err
}

func (e *CustomError) Code() ErrorCode {
	return e.code
}

func (e *CustomError) Unwrap() error {
	return e.Err
}

func (e *CustomError) Is(target error) bool {
	if t, ok := target.(*CustomError); ok {
		return e.code == t.code
	}
	return false
}

func (e *CustomError) PrintTrace() string {
	if e.Trace == "" {
		return ""
	}
	allTraces := append([]string{e.Trace}, e.PreviousTraces...)
	return strings.Join(allTraces, "\n")
}

func newCustomError(opts ...CustomErrorOption) *CustomError {
	e := &CustomError{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// New creates a new CustomError with the given message and code
func New(msg string, code ErrorCode) *CustomError {
	return newCustomError(WithMessage(msg), WithCode(code))
}

func wrapError(err error, opts ...CustomErrorOption) error {
	trace := captureStackTrace()
	var previousTraces []string
	if customErr, ok := err.(*CustomError); ok {
		previousTraces = append([]string{customErr.Trace}, customErr.PreviousTraces...)
		// 기존 에러 업데이트
		customErr.Trace = trace
		customErr.PreviousTraces = previousTraces
		for _, opt := range opts {
			opt(customErr)
		}
		return customErr
	}
	allOpts := append([]CustomErrorOption{WithTrace(trace), WithPreviousTraces(previousTraces), WithCause(err)}, opts...)
	return newCustomError(allOpts...)
}

func WrapTrace(err error) error {
	return wrapError(err, WithMessage("An error occurred"))
}

func WrapMessageWithCode(err error, errCode ErrorCode, msg string) error {
	return wrapError(err, WithMessage(msg), WithCode(errCode))
}

func WrapMessage(err error, msg string) error {
	return wrapError(err, WithMessage(msg), WithCode(ErrorInternalServer))
}

func captureStackTrace() string {
	var pcs [1]uintptr
	n := runtime.Callers(3, pcs[:]) // 3을 사용하여 호출자의 호출자에서 시작
	if n == 0 {
		return "unknown"
	}
	frame, _ := runtime.CallersFrames(pcs[:n]).Next()
	return fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function)
}

// The function "Cause" recursively retrieves the root cause of an error by checking if the error
// implements the CustomError interface.
func Cause(err error) error {
	if customErr, ok := err.(*CustomError); ok && customErr.Cause() != nil {
		return Cause(customErr.Cause())
	}
	return err
}

// The Unwrap function takes an error and return unwrapped error.
func Unwrap(err error) error {
	if customErr, ok := err.(*CustomError); ok && customErr != nil {
		return customErr.Cause()
	}
	return err
}

func Trace(err error) string {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.PrintTrace()
	}
	return ""
}

// IsCustomError checks if the error is a CustomError
func IsCustomError(err error) bool {
	_, ok := err.(*CustomError)
	return ok
}
