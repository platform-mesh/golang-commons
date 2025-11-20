package ratelimit

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/api"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/util"
	"github.com/platform-mesh/golang-commons/logger"
)

type RateLimiter struct {
	interval   time.Duration
	bypassFunc func(runtimeobject.RuntimeObject) bool
}

func NewRateLimiter(interval time.Duration, bypassFunc func(runtimeobject.RuntimeObject) bool) *RateLimiter {
	return &RateLimiter{
		interval:   interval,
		bypassFunc: bypassFunc,
	}
}

// ReconcileRequired should return true if a reconciling is required
func (r *RateLimiter) ReconcileRequired(instance runtimeobject.RuntimeObject, log *logger.Logger) bool {
	instanceStatusObj := util.MustToInterface[api.RuntimeObjectRateLimitStatus](instance, log)

	// Always reconcile if generation has changed
	if instance.GetGeneration() != instanceStatusObj.GetObservedGeneration() {
		return true
	}

	// Check if time has passed
	lastReconcile := instanceStatusObj.GetLastReconcileTime()
	nextFullReconcile := lastReconcile.Time.Add(r.interval)
	timeToNextReconcileHasPassed := v1.Now().After(nextFullReconcile)
	if lastReconcile.IsZero() {
		timeToNextReconcileHasPassed = true
	}

	if timeToNextReconcileHasPassed {
		return true
	}

	// Check bypass function
	if r.bypassFunc != nil && r.bypassFunc(instance) {
		log.Debug().Msg("Bypassing rate limit interval")
		return true
	}

	return false
}

func (r *RateLimiter) OnNextReconcile(instance runtimeobject.RuntimeObject, log *logger.Logger) (ctrl.Result, error) {
	instanceStatusObj := util.MustToInterface[api.RuntimeObjectRateLimitStatus](instance, log)
	nextFullReconcile := instanceStatusObj.GetLastReconcileTime().Time.Add(r.interval)
	until := time.Until(nextFullReconcile)

	log.Info().Time("next-full-reconcile", nextFullReconcile).Float64("until", until.Seconds()).Msg("Preventing too frequent reconciles, requeuing")

	return ctrl.Result{RequeueAfter: until}, nil
}

func (r *RateLimiter) SetLastReconcileTime(instanceStatusObj api.RuntimeObjectRateLimitStatus, log *logger.Logger) {
	log.Debug().Msg("Setting last reconcile time for the instance")
	instanceStatusObj.SetLastReconcileTime(v1.Now())
}
