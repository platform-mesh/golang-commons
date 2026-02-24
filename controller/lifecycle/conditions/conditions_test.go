package conditions

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	pmtesting "github.com/platform-mesh/golang-commons/controller/testSupport"
	"github.com/platform-mesh/golang-commons/logger"
)

// Test the setReady function with an empty array
func TestSetReady(t *testing.T) {

	t.Run("TestSetReady with empty array", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		cm := NewConditionManager()
		// When
		cm.SetInstanceConditionReady(&condition, metav1.ConditionTrue)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})

	t.Run("TestSetReady with existing condition", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{
			{Type: "test", Status: metav1.ConditionFalse},
		}

		// When
		cm.SetInstanceConditionReady(&condition, metav1.ConditionTrue)

		// Then
		assert.Equal(t, 2, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[1].Status)
	})
}

func TestSetUnknown(t *testing.T) {

	t.Run("TestSetUnknown with empty array", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{}

		// When
		cm.SetInstanceConditionUnknownIfNotSet(&condition)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	t.Run("TestSetUnknown with existing ready condition", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{
			{Type: ConditionReady, Status: metav1.ConditionTrue},
		}

		// When
		cm.SetInstanceConditionUnknownIfNotSet(&condition)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})
}

func TestSetSubroutineConditionToUnknownIfNotSet(t *testing.T) {
	log, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	unknownTests := []struct {
		Name         string
		WantsMessage string
		IsFinalize   bool
	}{
		{
			Name:         "TestSetSubroutineConditionToUnknownIfNotSet with empty array and finalize false",
			IsFinalize:   false,
			WantsMessage: "The subroutine is processing",
		},
		{
			Name:         "TestSetSubroutineConditionToUnknownIfNotSet with empty array and finalize true",
			IsFinalize:   true,
			WantsMessage: "The subroutine finalization is processing",
		},
	}
	for _, tt := range unknownTests {
		t.Run(tt.Name, func(t *testing.T) {
			// Given
			condition := []metav1.Condition{}
			cm := NewConditionManager()

			// When
			cm.SetSubroutineConditionToUnknownIfNotSet(
				&condition,
				pmtesting.ChangeStatusSubroutine{},
				tt.IsFinalize,
				log,
			)

			// Then
			assert.Equal(t, 1, len(condition))
			assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
			assert.Equal(t, tt.WantsMessage, condition[0].Message)
		})
	}

	t.Run("TestSetSubroutineConditionToUnknownIfNotSet with existing condition", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{
			{Type: "test", Status: metav1.ConditionFalse},
		}

		// When
		cm.SetSubroutineConditionToUnknownIfNotSet(&condition, pmtesting.ChangeStatusSubroutine{}, false, log)

		// Then
		assert.Equal(t, 2, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[1].Status)
	})

	t.Run("TestSetSubroutineConditionToUnknownIfNotSet with existing ready", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		subroutine := pmtesting.ChangeStatusSubroutine{}
		condition := []metav1.Condition{
			{Type: "test", Status: metav1.ConditionFalse},
			{Type: fmt.Sprintf("%s_Ready", subroutine.GetName()), Status: metav1.ConditionTrue},
		}

		// When
		cm.SetSubroutineConditionToUnknownIfNotSet(&condition, subroutine, false, log)

		// Then
		assert.Equal(t, 2, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[1].Status)
	})
}

func TestSubroutineCondition(t *testing.T) {
	log, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	// Add a test case to set a subroutine condition to ready if it was successfull
	t.Run("TestSetSubroutineConditionReady", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		subroutine := pmtesting.ChangeStatusSubroutine{}

		// When
		cm.SetSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, nil, false, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to unknown if it is still processing
	t.Run("TestSetSubroutineConditionProcessing", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		subroutine := pmtesting.ChangeStatusSubroutine{}

		// When
		cm.SetSubroutineCondition(
			&condition,
			subroutine,
			controllerruntime.Result{RequeueAfter: 1 * time.Second},
			nil,
			false,
			log,
		)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to false if it failed
	t.Run("TestSetSubroutineConditionError", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		cm := NewConditionManager()
		subroutine := pmtesting.ChangeStatusSubroutine{}

		// When
		cm.SetSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, errors.New("failed"), false, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
	})

	// Add a test case to set a subroutine condition for isFinalize true
	t.Run("TestSetSubroutineFinalizeConditionReady", func(t *testing.T) {
		// Given
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		subroutine := pmtesting.ChangeStatusSubroutine{}

		// When
		cm.SetSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, nil, true, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to unknown if it is still processing
	t.Run("TestSetSubroutineFinalizeConditionProcessing", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		cm := NewConditionManager()
		subroutine := pmtesting.ChangeStatusSubroutine{}

		// When
		cm.SetSubroutineCondition(
			&condition,
			subroutine,
			controllerruntime.Result{RequeueAfter: 1 * time.Second},
			nil,
			true,
			log,
		)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to false if it failed
	t.Run("TestSetSubroutineFinalizeConditionError", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		cm := NewConditionManager()
		subroutine := pmtesting.ChangeStatusSubroutine{}

		// When
		cm.SetSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, errors.New("failed"), true, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
	})
}

func TestSetSubroutineConditionFromResult(t *testing.T) {
	log, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	t.Run("Continue outcome sets ConditionTrue with Complete reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.OK()

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
		assert.Equal(t, "Complete", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "is complete")
	})

	t.Run("Continue outcome with requeue sets ConditionUnknown with Processing reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: 1 * time.Second})

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
		assert.Equal(t, "Processing", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "is processing")
	})

	t.Run("Skipped outcome sets ConditionTrue with Skipped reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Skip("no work to do")

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
		assert.Equal(t, "Skipped", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "was skipped: no work to do")
	})

	t.Run("Skipped outcome without reason sets ConditionTrue with Skipped reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Result{Outcome: subroutine.Skipped}

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
		assert.Equal(t, "Skipped", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "was skipped")
		assert.NotContains(t, condition[0].Message, ":")
	})

	t.Run("StopChain outcome sets ConditionTrue with StopChain reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Stop("resource is being deleted")

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
		assert.Equal(t, "StopChain", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "stopped chain: resource is being deleted")
	})

	t.Run("ErrorRetry outcome sets ConditionFalse with ErrorRetry reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Retry(errors.New("temporary API failure"))

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
		assert.Equal(t, "ErrorRetry", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "retryable error")
		assert.Contains(t, condition[0].Message, "temporary API failure")
	})

	t.Run("ErrorContinue outcome sets ConditionFalse with ErrorContinue reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Fail(errors.New("optional feature unavailable"))

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
		assert.Equal(t, "ErrorContinue", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "non-blocking error")
		assert.Contains(t, condition[0].Message, "optional feature unavailable")
	})

	t.Run("ErrorStop outcome sets ConditionFalse with ErrorStop reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.StopWithError(errors.New("invalid configuration"), "config validation failed")

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
		assert.Equal(t, "ErrorStop", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "terminal error")
		assert.Contains(t, condition[0].Message, "invalid configuration")
	})

	t.Run("Finalize mode uses correct condition name format", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Skip("no finalizer to remove")

		cm.SetSubroutineConditionFromResult(&condition, sub, result, true, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, "changeStatus_Finalize", condition[0].Type)
		assert.Equal(t, "Skipped", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "subroutine finalization was skipped")
	})

	t.Run("StopChain outcome without reason sets ConditionTrue with StopChain reason", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Result{Outcome: subroutine.StopChain}

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
		assert.Equal(t, "StopChain", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "stopped chain")
		assert.NotContains(t, condition[0].Message, ":")
	})

	t.Run("Unknown outcome defaults to Processing", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.Result{Outcome: subroutine.Outcome(999)} // invalid outcome

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
		assert.Equal(t, "Processing", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "is processing")
	})

	t.Run("Updates existing condition rather than appending", func(t *testing.T) {
		cm := NewConditionManager()
		sub := pmtesting.ChangeStatusSubroutine{}
		condition := []metav1.Condition{
			{
				Type:    "changeStatus_Ready",
				Status:  metav1.ConditionUnknown,
				Reason:  "Processing",
				Message: "The subroutine is processing",
			},
		}

		result := subroutine.OK()
		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
		assert.Equal(t, "Complete", condition[0].Reason)
	})

	t.Run("RetryWithRequeue sets ErrorRetry with custom requeue time", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}
		result := subroutine.RetryWithRequeue(
			controllerruntime.Result{RequeueAfter: 5 * time.Minute},
			errors.New("rate limited"),
		)

		cm.SetSubroutineConditionFromResult(&condition, sub, result, false, log)

		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
		assert.Equal(t, "ErrorRetry", condition[0].Reason)
		assert.Contains(t, condition[0].Message, "rate limited")
	})

	t.Run("Returns true when condition changed", func(t *testing.T) {
		cm := NewConditionManager()
		condition := []metav1.Condition{}
		sub := pmtesting.ChangeStatusSubroutine{}

		changed := cm.SetSubroutineConditionFromResult(&condition, sub, subroutine.OK(), false, log)

		assert.True(t, changed)
	})

	t.Run("Returns false when condition unchanged", func(t *testing.T) {
		cm := NewConditionManager()
		sub := pmtesting.ChangeStatusSubroutine{}
		condition := []metav1.Condition{}

		// Set initial condition
		cm.SetSubroutineConditionFromResult(&condition, sub, subroutine.OK(), false, log)

		// Set same condition again
		changed := cm.SetSubroutineConditionFromResult(&condition, sub, subroutine.OK(), false, log)

		assert.False(t, changed)
	})
}
