package ratelimiter

import (
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
)

type StaticThenExponentialRateLimiter[T comparable] struct {
	failuresLock sync.Mutex
	firstAttempt map[T]time.Time

	staticDelay  time.Duration
	staticWindow time.Duration

	exponential workqueue.TypedRateLimiter[T]
	clock       clock.Clock
}

func NewStaticThenExponentialRateLimiter[T comparable](cfg Config) *StaticThenExponentialRateLimiter[T] {
	return &StaticThenExponentialRateLimiter[T]{
		staticDelay:  cfg.StaticRequeueDelay,
		staticWindow: cfg.StaticWindow,
		exponential: workqueue.NewTypedItemExponentialFailureRateLimiter[T](
			cfg.ExponentialInitialBackoff,
			cfg.ExponentialMaxBackoff,
		),
		firstAttempt: make(map[T]time.Time),
		clock:        clock.RealClock{},
	}
}

func (r *StaticThenExponentialRateLimiter[T]) When(item T) time.Duration {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	now := r.clock.Now()

	first, exists := r.firstAttempt[item]
	if !exists {
		first = now
		r.firstAttempt[item] = first
	}

	timeSinceFirst := now.Sub(first)
	if timeSinceFirst <= r.staticWindow {
		return r.staticDelay
	}

	return r.exponential.When(item)
}

func (r *StaticThenExponentialRateLimiter[T]) Forget(item T) {
	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	delete(r.firstAttempt, item)
	r.exponential.Forget(item)
}

func (r *StaticThenExponentialRateLimiter[T]) NumRequeues(item T) int {
	return r.exponential.NumRequeues(item)
}
