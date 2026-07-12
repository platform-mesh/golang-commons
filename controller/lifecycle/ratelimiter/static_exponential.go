package ratelimiter

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/platform-mesh/golang-commons/logger"

	"k8s.io/client-go/util/workqueue"
	"k8s.io/utils/clock"
)

type StaticThenExponentialRateLimiter[T comparable] struct {
	failuresLock   sync.RWMutex
	staticAttempts map[T]time.Time

	staticDelay  time.Duration
	staticWindow time.Duration

	exponential workqueue.TypedRateLimiter[T]
	clock       clock.Clock
	logger      *logger.Logger
}

func NewStaticThenExponentialRateLimiter[T comparable](cfg Config) (*StaticThenExponentialRateLimiter[T], error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &StaticThenExponentialRateLimiter[T]{
		staticDelay:  cfg.StaticRequeueDelay,
		staticWindow: cfg.StaticWindow,
		exponential: workqueue.NewTypedItemExponentialFailureRateLimiter[T](
			cfg.ExponentialInitialBackoff,
			cfg.ExponentialMaxBackoff,
		),
		staticAttempts: make(map[T]time.Time),
		clock:          clock.RealClock{},
	}, nil
}

func (r *StaticThenExponentialRateLimiter[T]) SetLogger(logger *logger.Logger) {
	r.logger = logger.MustChildLoggerWithAttributes("component", "StaticThenExponentialRateLimiter")
}

func (r *StaticThenExponentialRateLimiter[T]) When(item T) time.Duration {
	now := r.clock.Now()

	r.failuresLock.RLock()
	first, exists := r.staticAttempts[item]
	expFailures := r.exponential.NumRequeues(item)
	mapSize := len(r.staticAttempts)
	r.failuresLock.RUnlock()

	if !exists {
		r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] STATE_RESET_DETECTED: item=%v, map_size=%d, exp_failures=%d",
			item, mapSize, expFailures))

		r.failuresLock.Lock()
		r.staticAttempts[item] = now
		r.failuresLock.Unlock()

		r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] INIT_STATIC: item=%v, static_delay=%v, timestamp=%v",
			item, r.staticDelay, now.Unix()))

		return r.staticDelay
	}

	timeSinceFirst := now.Sub(first)

	r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] TRACKING_ACTIVE: item=%v, time_since_first=%v, static_window=%v, exp_failures=%d, in_static=%v",
		item, timeSinceFirst, r.staticWindow, expFailures, timeSinceFirst <= r.staticWindow))

	if timeSinceFirst <= r.staticWindow {
		r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] STATIC_PHASE: item=%v, delay=%v, elapsed=%v",
			item, r.staticDelay, timeSinceFirst))
		return r.staticDelay
	}

	delay := r.exponential.When(item)
	updatedExpFailures := r.exponential.NumRequeues(item)

	r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] EXPONENTIAL_PHASE: item=%v, delay=%v, exp_failures=%d, time_since_first=%v",
		item, delay, updatedExpFailures, timeSinceFirst))

	if delay >= 1*time.Hour {
		r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] MAX_BACKOFF_REACHED: item=%v, delay=%v, exp_failures=%d, time_since_first=%v",
			item, delay, updatedExpFailures, timeSinceFirst))
	}

	return delay
}

func (r *StaticThenExponentialRateLimiter[T]) Forget(item T) {
	r.failuresLock.RLock()
	_, existsInStatic := r.staticAttempts[item]
	staticMapSize := len(r.staticAttempts)
	expFailuresBefore := r.exponential.NumRequeues(item)
	r.failuresLock.RUnlock()

	// Get caller info
	pc, file, line, ok := runtime.Caller(1)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d (%s)", file, line, runtime.FuncForPC(pc).Name())
	}

	r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] FORGET_CALLED: item=%v, caller=%s, exists_in_static=%v, map_size=%d, exp_failures_before=%d",
		item, caller, existsInStatic, staticMapSize, expFailuresBefore))

	r.failuresLock.Lock()
	defer r.failuresLock.Unlock()

	if first, exists := r.staticAttempts[item]; exists {
		timeSinceFirst := r.clock.Now().Sub(first)
		r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] FORGET_DELETE_STATIC: item=%v, first_time=%v, tracked_duration=%v",
			item, first.Unix(), timeSinceFirst))
	}

	delete(r.staticAttempts, item)
	r.exponential.Forget(item)

	expFailuresAfter := r.exponential.NumRequeues(item)
	staticMapSizeAfter := len(r.staticAttempts)

	r.logger.Info().Msg(fmt.Sprintf("[RateLimiter] FORGET_COMPLETED: item=%v, map_size_after=%d, exp_failures_after=%d, was_in_error=%v",
		item, staticMapSizeAfter, expFailuresAfter, expFailuresBefore > 0))
}

func (r *StaticThenExponentialRateLimiter[T]) NumRequeues(item T) int {
	return r.exponential.NumRequeues(item)
}
