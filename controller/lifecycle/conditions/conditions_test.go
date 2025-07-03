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

	pmtesting "github.com/platform-mesh/golang-commons/controller/lifecycle/testing"
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
			cm.SetSubroutineConditionToUnknownIfNotSet(&condition, pmtesting.ChangeStatusSubroutine{}, tt.IsFinalize, log)

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
		cm.SetSubroutineCondition(&condition, subroutine, controllerruntime.Result{RequeueAfter: 1 * time.Second}, nil, false, log)

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
		cm.SetSubroutineCondition(&condition, subroutine, controllerruntime.Result{RequeueAfter: 1 * time.Second}, nil, true, log)

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
