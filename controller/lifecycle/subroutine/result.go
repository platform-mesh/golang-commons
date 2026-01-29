package subroutine

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type Outcome int

const (
	Continue      Outcome = iota // proceed to next subroutine
	StopChain                    // stop processing remaining subroutines
	Skipped                      // subroutine was skipped
	ErrorRetry                   // error -> requeue
	ErrorContinue                // error, but continue processing
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
	default:
		return "Unknown"
	}
}

type Result struct {
	Ctrl    ctrl.Result
	Outcome Outcome
	Error   error
	Reason  string
	Sentry  bool
}

func OK() Result {
	return Result{Outcome: Continue}
}

func OKWithRequeue(r ctrl.Result) Result {
	return Result{Ctrl: r, Outcome: Continue}
}

func Stop(reason string) Result {
	return Result{Outcome: StopChain, Reason: reason}
}

func StopWithRequeue(r ctrl.Result, reason string) Result {
	return Result{Ctrl: r, Outcome: StopChain, Reason: reason}
}

func Skip(reason string) Result {
	return Result{Outcome: Skipped, Reason: reason}
}

func Retry(err error) Result {
	return Result{Outcome: ErrorRetry, Error: err, Sentry: true}
}

func RetryWithResult(r ctrl.Result, err error) Result {
	return Result{Ctrl: r, Outcome: ErrorRetry, Error: err, Sentry: true}
}

func RetryNoSentry(err error) Result {
	return Result{Outcome: ErrorRetry, Error: err}
}

func Fail(err error) Result {
	return Result{Outcome: ErrorContinue, Error: err, Sentry: true}
}

func FailNoSentry(err error) Result {
	return Result{Outcome: ErrorContinue, Error: err}
}

func FailWithReason(err error, reason string) Result {
	return Result{Outcome: ErrorContinue, Error: err, Reason: reason, Sentry: true}
}

func (r Result) IsError() bool {
	return r.Outcome == ErrorRetry || r.Outcome == ErrorContinue
}

func (r Result) ShouldContinue() bool {
	return r.Outcome == Continue || r.Outcome == Skipped || r.Outcome == ErrorContinue
}

func (r Result) ShouldStop() bool {
	return r.Outcome == StopChain || r.Outcome == ErrorRetry
}

func (r Result) ShouldRequeue() bool {
	return r.Outcome == ErrorRetry || r.Ctrl.RequeueAfter > 0 || r.Ctrl.Requeue
}
