package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/logger"
)

// MockRuntimeObject implements RuntimeObject and RuntimeObjectRateLimitStatus
type MockRuntimeObject struct {
	mock.Mock
	runtimeobject.RuntimeObject
}

func (m *MockRuntimeObject) GetGeneration() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockRuntimeObject) GetObservedGeneration() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockRuntimeObject) GetLastReconcileTime() v1.Time {
	args := m.Called()
	return args.Get(0).(v1.Time)
}

func (m *MockRuntimeObject) SetLastReconcileTime(t v1.Time) {
	m.Called(t)
}

func (m *MockRuntimeObject) DeepCopyObject() runtime.Object {
	args := m.Called()
	return args.Get(0).(runtime.Object)
}

func TestRateLimiter_ReconcileRequired(t *testing.T) {
	log := logger.StdLogger
	interval := 5 * time.Minute

	tests := []struct {
		name               string
		generation         int64
		observedGeneration int64
		lastReconcileTime  time.Time
		bypassResult       bool
		expectRequired     bool
	}{
		{
			name:               "Generation changed",
			generation:         2,
			observedGeneration: 1,
			lastReconcileTime:  time.Now(),
			expectRequired:     true,
		},
		{
			name:               "Time passed",
			generation:         1,
			observedGeneration: 1,
			lastReconcileTime:  time.Now().Add(-10 * time.Minute),
			expectRequired:     true,
		},
		{
			name:               "Time not passed",
			generation:         1,
			observedGeneration: 1,
			lastReconcileTime:  time.Now().Add(-1 * time.Minute),
			expectRequired:     false,
		},
		{
			name:               "Zero LastReconcileTime",
			generation:         1,
			observedGeneration: 1,
			lastReconcileTime:  time.Time{},
			expectRequired:     true,
		},
		{
			name:               "Bypass true",
			generation:         1,
			observedGeneration: 1,
			lastReconcileTime:  time.Now().Add(-1 * time.Minute),
			bypassResult:       true,
			expectRequired:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockObj := new(MockRuntimeObject)
			mockObj.On("GetGeneration").Return(tt.generation)
			mockObj.On("GetObservedGeneration").Return(tt.observedGeneration)
			mockObj.On("GetLastReconcileTime").Return(v1.NewTime(tt.lastReconcileTime))

			bypassFunc := func(obj runtimeobject.RuntimeObject) bool {
				return tt.bypassResult
			}

			rl := NewRateLimiter(interval, bypassFunc)
			required := rl.ReconcileRequired(mockObj, log)

			assert.Equal(t, tt.expectRequired, required)
		})
	}
}

func TestRateLimiter_OnNextReconcile(t *testing.T) {
	log := logger.StdLogger
	interval := 5 * time.Minute
	lastReconcile := time.Now().Add(-2 * time.Minute)

	mockObj := new(MockRuntimeObject)
	mockObj.On("GetLastReconcileTime").Return(v1.NewTime(lastReconcile))

	rl := NewRateLimiter(interval, nil)
	result, err := rl.OnNextReconcile(mockObj, log)

	assert.NoError(t, err)
	// Expected requeue is approx 3 minutes (5 - 2)
	expectedRequeue := 3 * time.Minute
	// Allow 1 second delta
	assert.InDelta(t, expectedRequeue.Seconds(), result.RequeueAfter.Seconds(), 1.0)
}

func TestRateLimiter_SetLastReconcileTime(t *testing.T) {
	log := logger.StdLogger
	mockObj := new(MockRuntimeObject)
	mockObj.On("SetLastReconcileTime", mock.AnythingOfType("v1.Time")).Run(func(args mock.Arguments) {
		val := args.Get(0).(v1.Time)
		assert.WithinDuration(t, time.Now(), val.Time, 1*time.Second)
	})

	rl := NewRateLimiter(0, nil)
	rl.SetLastReconcileTime(mockObj, log)

	mockObj.AssertExpectations(t)
}
