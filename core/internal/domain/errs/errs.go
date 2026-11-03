package errs

import "errors"

var (
	ErrValidationError      = errors.New("validation error")
	ErrUnauthenticated      = errors.New("invalid login or password")
	ErrRunNotFound          = errors.New("run not found")
	ErrRunNotAwaitingReview = errors.New("run is not awaiting review")
	ErrRunHasNoImage        = errors.New("run has no image")
	ErrCaseNotFound         = errors.New("case not found")
	ErrBenchmarkNotFound    = errors.New("benchmark not found")
	ErrBenchmarkRunning     = errors.New("a benchmark is already running")
)

// ValidationError marks input that failed validation. It unwraps to the cause so the
// HTTP layer can name the offending field, and matches ErrValidationError in errors.Is.
type ValidationError struct {
	cause error
}

func NewValidationError(cause error) error {
	return &ValidationError{cause: cause}
}

func (e *ValidationError) Error() string {
	return e.cause.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.cause
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidationError
}
