package subroutine

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/errors"
)

// Outcome represents the result of a subroutine execution
// and determines how the lifecycle should proceed
type Outcome int

const (
	// Continue proceeds to the next subroutine in the chain
	// Use when the subroutine completed successfully
	Continue Outcome = iota

	// StopChain stops processing remaining subroutines without error
	// Use when further processing is unnecessary (e.g. resource is being deleted)
	StopChain

	// Skipped indicates the subroutine had nothing to do
	// Chain continues, useful for conditional logic (e.g. no finalizer to remove)
	Skipped

	// ErrorRetry indicates a transient error that should be retried
	// Stops the chain and requeues (e.g. temporary API failure)
	ErrorRetry

	// ErrorContinue indicates an error that should not block the chain
	// Logs the error but continues processing (e.g. optional feature unavailable)
	ErrorContinue

	// ErrorStop indicates a terminal error with no retry.
	// Use for permanent failures (e.g. invalid configuration)
	ErrorStop
)

func (o Outcome) String() string {
	switch o {
	case Continue:
		return "Continue"
	case StopChain:
		return "StopChain"
	case Skipped:
		return "Skipped"
	case ErrorRetry:
		return "ErrorRetry"
	case ErrorContinue:
		return "ErrorContinue"
	case ErrorStop:
		return "ErrorStop"
	default:
		return "Unknown"
	}
}

// Result is the return type for ChainSubroutine implementations
type Result struct {
	Ctrl    ctrl.Result
	Outcome Outcome
	Error   error
	Reason  string
	Sentry  bool
}

// WithSentry enables Sentry error reporting for this result
func (r Result) WithSentry() Result {
	r.Sentry = true
	return r
}

// WithReason attaches a human-readable explanation to the result
func (r Result) WithReason(reason string) Result {
	r.Reason = reason
	return r
}

// WithRequeue sets custom requeue timing for the result
func (r Result) WithRequeue(requeue ctrl.Result) Result {
	r.Ctrl = requeue
	return r
}

// OK indicates successful completion, proceeds to next subroutine
func OK() Result {
	return Result{Outcome: Continue}
}

// OKWithRequeue indicates successful completion with scheduled requeue
func OKWithRequeue(r ctrl.Result) Result {
	return Result{Ctrl: r, Outcome: Continue}
}

// Stop halts the chain without error (e.g., resource is being deleted).
func Stop(reason string) Result {
	return Result{Outcome: StopChain, Reason: reason}
}

// StopWithRequeue halts the chain and schedules a requeue
func StopWithRequeue(r ctrl.Result, reason string) Result {
	return Result{Ctrl: r, Outcome: StopChain, Reason: reason}
}

// Skip indicates the subroutine was not applicable
func Skip(reason string) Result {
	return Result{Outcome: Skipped, Reason: reason}
}

// Retry indicates a transient error that should be retried (e.g., API temporarily unavailable)
// NOTE: Use WithSentry() for unexpected errors
func Retry(err error) Result {
	return Result{Outcome: ErrorRetry, Error: err}
}

// RetryWithRequeue indicates a transient error with custom requeue timing
func RetryWithRequeue(r ctrl.Result, err error) Result {
	return Result{Ctrl: r, Outcome: ErrorRetry, Error: err}
}

// Fail indicates a non-critical error that should not stop the chain
// The error is logged but processing continues
func Fail(err error) Result {
	return Result{Outcome: ErrorContinue, Error: err}
}

// FailWithReason indicates a non-critical error with additional context
func FailWithReason(err error, reason string) Result {
	return Result{Outcome: ErrorContinue, Error: err, Reason: reason}
}

// StopWithError indicates a terminal error that should not be retried
func StopWithError(err error, reason string) Result {
	return Result{Outcome: ErrorStop, Error: err, Reason: reason}
}

// IsError returns true if the result represents an error condition
func (r Result) IsError() bool {
	return r.Outcome == ErrorRetry || r.Outcome == ErrorContinue || r.Outcome == ErrorStop
}

// ShouldContinue returns true if the chain should proceed to the next subroutine
func (r Result) ShouldContinue() bool {
	return r.Outcome == Continue || r.Outcome == Skipped || r.Outcome == ErrorContinue
}

// ShouldStop returns true if the chain should stop processing
func (r Result) ShouldStop() bool {
	return r.Outcome == StopChain || r.Outcome == ErrorRetry || r.Outcome == ErrorStop
}

// ShouldRequeue returns true if the reconciliation should be requeued
func (r Result) ShouldRequeue() bool {
	return r.Outcome == ErrorRetry || r.Ctrl.RequeueAfter > 0
}

// ToSubroutine converts Result to the legacy (ctrl.Result, OperatorError) format
// Used internally by ChainAdapter for backwards compatibility
func (r Result) ToSubroutine() (ctrl.Result, errors.OperatorError) {
	if r.Error != nil {
		retry := r.Outcome == ErrorRetry
		return r.Ctrl, errors.NewOperatorError(r.Error, retry, r.Sentry)
	}
	return r.Ctrl, nil
}
