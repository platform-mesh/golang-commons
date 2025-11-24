package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	clocktesting "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestStaticThenExponentialRateLimiter_Forget(t *testing.T) {
	cfg := Config{
		StaticRequeueDelay:        1 * time.Second,
		StaticWindow:              5 * time.Second,
		ExponentialInitialBackoff: 2 * time.Second,
		ExponentialMaxBackoff:     1 * time.Minute,
	}
	limiter := NewStaticThenExponentialRateLimiter[reconcile.Request](cfg)
	fakeClock := clocktesting.NewFakeClock(time.Now())
	limiter.clock = fakeClock

	item := reconcile.Request{NamespacedName: types.NamespacedName{Name: "name", Namespace: "namespace"}}
	require.Equal(t, cfg.StaticRequeueDelay, limiter.When(item))
	fakeClock.Step(10 * time.Second)
	_ = limiter.When(item)

	limiter.Forget(item)
	require.Equal(t, 0, limiter.NumRequeues(item))
	delay := limiter.When(item)
	require.Equal(t, cfg.StaticRequeueDelay, delay)
}
