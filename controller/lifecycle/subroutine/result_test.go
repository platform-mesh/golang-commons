package subroutine

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestOutcomeString(t *testing.T) {
	tests := []struct {
		name     string
		outcome  Outcome
		expected string
	}{
		{"Continue", Continue, "Continue"},
		{"StopChain", StopChain, "StopChain"},
		{"Skipped", Skipped, "Skipped"},
		{"ErrorRetry", ErrorRetry, "ErrorRetry"},
		{"ErrorContinue", ErrorContinue, "ErrorContinue"},
		{"ErrorStop", ErrorStop, "ErrorStop"},
		{"Unknown", Outcome(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.outcome.String())
		})
	}
}

func TestOK(t *testing.T) {
	result := OK()

	assert.Equal(t, Continue, result.Outcome)
	assert.Equal(t, ctrl.Result{}, result.Ctrl)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.Reason)
	assert.False(t, result.Sentry)
}

func TestOKWithRequeue(t *testing.T) {
	ctrlResult := ctrl.Result{RequeueAfter: 5 * time.Second}
	result := OKWithRequeue(ctrlResult)

	assert.Equal(t, Continue, result.Outcome)
	assert.Equal(t, ctrlResult, result.Ctrl)
	assert.Nil(t, result.Error)
}

func TestStop(t *testing.T) {
	reason := "dependency not ready"
	result := Stop(reason)

	assert.Equal(t, StopChain, result.Outcome)
	assert.Equal(t, reason, result.Reason)
	assert.Equal(t, ctrl.Result{}, result.Ctrl)
}

func TestStopWithRequeue(t *testing.T) {
	ctrlResult := ctrl.Result{RequeueAfter: 10 * time.Second}
	reason := "waiting for external resource"
	result := StopWithRequeue(ctrlResult, reason)

	assert.Equal(t, StopChain, result.Outcome)
	assert.Equal(t, reason, result.Reason)
	assert.Equal(t, ctrlResult, result.Ctrl)
}

func TestSkip(t *testing.T) {
	reason := "no finalizer"
	result := Skip(reason)

	assert.Equal(t, Skipped, result.Outcome)
	assert.Equal(t, reason, result.Reason)
}

func TestRetry(t *testing.T) {
	err := errors.New("temporary failure")
	result := Retry(err).WithSentry()

	assert.Equal(t, ErrorRetry, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.True(t, result.Sentry)
}

func TestRetryNoSentry(t *testing.T) {
	err := errors.New("expected error")
	result := Retry(err)

	assert.Equal(t, ErrorRetry, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.False(t, result.Sentry)
}

func TestRetryWithRequeue(t *testing.T) {
	err := errors.New("temporary failure")
	ctrlResult := ctrl.Result{RequeueAfter: 30 * time.Second}
	result := RetryWithRequeue(ctrlResult, err).WithSentry()

	assert.Equal(t, ErrorRetry, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, ctrlResult, result.Ctrl)
	assert.True(t, result.Sentry)
}

func TestFail(t *testing.T) {
	err := errors.New("permanent failure")
	result := Fail(err).WithSentry()

	assert.Equal(t, ErrorContinue, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.True(t, result.Sentry)
}

func TestFailNoSentry(t *testing.T) {
	err := errors.New("expected permanent failure")
	result := Fail(err)

	assert.Equal(t, ErrorContinue, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.False(t, result.Sentry)
}

func TestFailWithReason(t *testing.T) {
	err := errors.New("configuration error")
	reason := "invalid configuration"
	result := FailWithReason(err, reason).WithSentry()

	assert.Equal(t, ErrorContinue, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, reason, result.Reason)
	assert.True(t, result.Sentry)
}

func TestStopWithError(t *testing.T) {
	err := errors.New("fatal error")
	reason := "cannot proceed"
	result := StopWithError(err, reason).WithSentry()

	assert.Equal(t, ErrorStop, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, reason, result.Reason)
	assert.True(t, result.Sentry)
}

func TestWithReason(t *testing.T) {
	result := Retry(errors.New("err")).WithReason("custom reason")

	assert.Equal(t, "custom reason", result.Reason)
	assert.Equal(t, ErrorRetry, result.Outcome)
}

func TestWithRequeue(t *testing.T) {
	ctrlResult := ctrl.Result{RequeueAfter: 5 * time.Minute}
	result := Fail(errors.New("err")).WithRequeue(ctrlResult)

	assert.Equal(t, ctrlResult, result.Ctrl)
	assert.Equal(t, ErrorContinue, result.Outcome)
}

func TestChainMethods(t *testing.T) {
	// Test chaining multiple methods together
	ctrlResult := ctrl.Result{RequeueAfter: 10 * time.Second}
	err := errors.New("test error")

	result := Retry(err).
		WithSentry().
		WithReason("chained reason").
		WithRequeue(ctrlResult)

	assert.Equal(t, ErrorRetry, result.Outcome)
	assert.Equal(t, err, result.Error)
	assert.True(t, result.Sentry)
	assert.Equal(t, "chained reason", result.Reason)
	assert.Equal(t, ctrlResult, result.Ctrl)
}

func TestIsError(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"Continue is not error", OK(), false},
		{"OKWithRequeue is not error", OKWithRequeue(ctrl.Result{RequeueAfter: time.Second}), false},
		{"Stop is not error", Stop("reason"), false},
		{"StopWithRequeue is not error", StopWithRequeue(ctrl.Result{RequeueAfter: time.Second}, "reason"), false},
		{"Skip is not error", Skip("reason"), false},
		{"Retry is error", Retry(errors.New("err")).WithSentry(), true},
		{"RetryNoSentry is error", Retry(errors.New("err")), true},
		{"Fail is error", Fail(errors.New("err")).WithSentry(), true},
		{"FailNoSentry is error", Fail(errors.New("err")), true},
		{"FailWithReason is error", FailWithReason(errors.New("err"), "reason").WithSentry(), true},
		{"ErrorStop is error", StopWithError(errors.New("err"), "reason").WithSentry(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.IsError())
		})
	}
}

func TestShouldContinue(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"Continue should continue", OK(), true},
		{"OKWithRequeue should continue", OKWithRequeue(ctrl.Result{RequeueAfter: time.Second}), true},
		{"Stop should not continue", Stop("reason"), false},
		{
			"StopWithRequeue should not continue",
			StopWithRequeue(ctrl.Result{RequeueAfter: time.Second}, "reason"),
			false,
		},
		{"Skip should continue", Skip("reason"), true},
		{"Retry should not continue", Retry(errors.New("err")).WithSentry(), false},
		{"Fail should continue", Fail(errors.New("err")).WithSentry(), true},
		{"FailNoSentry should continue", Fail(errors.New("err")), true},
		{"ErrorStop should not continue", StopWithError(errors.New("err"), "reason").WithSentry(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.ShouldContinue())
		})
	}
}

func TestShouldStop(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"Continue should not stop", OK(), false},
		{"OKWithRequeue should not stop", OKWithRequeue(ctrl.Result{RequeueAfter: time.Second}), false},
		{"Stop should stop", Stop("reason"), true},
		{"StopWithRequeue should stop", StopWithRequeue(ctrl.Result{RequeueAfter: time.Second}, "reason"), true},
		{"Skip should not stop", Skip("reason"), false},
		{"Retry should stop", Retry(errors.New("err")).WithSentry(), true},
		{"Fail should not stop", Fail(errors.New("err")).WithSentry(), false},
		{"ErrorStop should stop", StopWithError(errors.New("err"), "reason").WithSentry(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.ShouldStop())
		})
	}
}

func TestShouldRequeue(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"Continue should not requeue", OK(), false},
		{"OKWithRequeue should requeue", OKWithRequeue(ctrl.Result{RequeueAfter: time.Second}), true},
		{"Stop should not requeue", Stop("reason"), false},
		{"StopWithRequeue should requeue", StopWithRequeue(ctrl.Result{RequeueAfter: time.Second}, "reason"), true},
		{"Skip should not requeue", Skip("reason"), false},
		{"Retry should requeue", Retry(errors.New("err")).WithSentry(), true},
		{"Fail should not requeue", Fail(errors.New("err")).WithSentry(), false},
		{"ErrorStop should not requeue", StopWithError(errors.New("err"), "reason").WithSentry(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.ShouldRequeue())
		})
	}
}

func TestToV1(t *testing.T) {
	t.Run("OK result converts to nil error", func(t *testing.T) {
		result := OK()
		ctrlResult, err := result.ToSubroutine()

		assert.Equal(t, ctrl.Result{}, ctrlResult)
		assert.Nil(t, err)
	})

	t.Run("Retry result converts to retry error", func(t *testing.T) {
		testErr := errors.New("test error")
		result := Retry(testErr).WithSentry()
		ctrlResult, err := result.ToSubroutine()

		assert.Equal(t, ctrl.Result{}, ctrlResult)
		assert.NotNil(t, err)
		assert.Equal(t, testErr, err.Err())
		assert.True(t, err.Retry())
		assert.True(t, err.Sentry())
	})

	t.Run("Fail result converts to non-retry error", func(t *testing.T) {
		testErr := errors.New("test error")
		result := Fail(testErr)
		ctrlResult, err := result.ToSubroutine()

		assert.Equal(t, ctrl.Result{}, ctrlResult)
		assert.NotNil(t, err)
		assert.Equal(t, testErr, err.Err())
		assert.False(t, err.Retry())
		assert.False(t, err.Sentry())
	})
}
