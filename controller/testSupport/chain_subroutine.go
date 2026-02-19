package testSupport

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
)

const ChainSubroutineFinalizer = "finalizer-chain"

type OKChainSubroutine struct {
	Name string
}

func (s OKChainSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	return subroutine.OK()
}

func (s OKChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	return subroutine.OK()
}

func (s OKChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "OKChainSubroutine"
}

func (s OKChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{ChainSubroutineFinalizer}
}

type StopChainSubroutine struct {
	Name         string
	Reason       string
	RequeueAfter time.Duration
}

func (s StopChainSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	if s.RequeueAfter > 0 {
		return subroutine.StopWithRequeue(ctrl.Result{RequeueAfter: s.RequeueAfter}, s.Reason)
	}
	return subroutine.Stop(s.Reason)
}

func (s StopChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	return subroutine.OK()
}

func (s StopChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "StopChainSubroutine"
}

func (s StopChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return nil
}

type RetryChainSubroutine struct {
	Name   string
	Err    error
	Sentry bool
}

func (s RetryChainSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	result := subroutine.Retry(s.Err)
	if s.Sentry {
		result = result.WithSentry()
	}
	return result
}

func (s RetryChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	result := subroutine.Retry(s.Err)
	if s.Sentry {
		result = result.WithSentry()
	}
	return result
}

func (s RetryChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "RetryChainSubroutine"
}

func (s RetryChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{ChainSubroutineFinalizer}
}

type FailChainSubroutine struct {
	Name   string
	Err    error
	Sentry bool
	Reason string
}

func (s FailChainSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	var result subroutine.Result
	if s.Reason != "" {
		result = subroutine.FailWithReason(s.Err, s.Reason)
	} else {
		result = subroutine.Fail(s.Err)
	}
	if s.Sentry {
		result = result.WithSentry()
	}
	return result
}

func (s FailChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	return subroutine.OK()
}

func (s FailChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "FailChainSubroutine"
}

func (s FailChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return nil
}

type ChangeStatusChainSubroutine struct {
	Client client.Client
	Name   string
}

func (s ChangeStatusChainSubroutine) Process(
	_ context.Context,
	runtimeObj runtimeobject.RuntimeObject,
) subroutine.Result {
	if instance, ok := runtimeObj.(*TestApiObject); ok {
		instance.Status.Some = "changed by chain"
	}
	if instance, ok := runtimeObj.(*ImplementingSpreadReconciles); ok {
		instance.Status.Some = "changed by chain"
	}
	if instance, ok := runtimeObj.(*ImplementConditions); ok {
		instance.Status.Some = "changed by chain"
	}
	return subroutine.OK()
}

func (s ChangeStatusChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	return subroutine.OK()
}

func (s ChangeStatusChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "ChangeStatusChainSubroutine"
}

func (s ChangeStatusChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{"changestatus-chain"}
}

type FinalizerChainSubroutine struct {
	Client       client.Client
	Err          error
	RequeueAfter time.Duration
	Name         string
}

func (s FinalizerChainSubroutine) Process(_ context.Context, runtimeObj runtimeobject.RuntimeObject) subroutine.Result {
	instance := runtimeObj.(*TestApiObject)
	instance.Status.Some = "processed by chain"
	return subroutine.OK()
}

func (s FinalizerChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	if s.Err != nil {
		return subroutine.Retry(s.Err).WithSentry()
	}
	if s.RequeueAfter > 0 {
		return subroutine.OKWithRequeue(ctrl.Result{RequeueAfter: s.RequeueAfter})
	}
	return subroutine.OK()
}

func (s FinalizerChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "FinalizerChainSubroutine"
}

func (s FinalizerChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{ChainSubroutineFinalizer}
}

type ErrorStopChainSubroutine struct {
	Name   string
	Err    error
	Reason string
	Sentry bool
}

func (s ErrorStopChainSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	result := subroutine.StopWithError(s.Err, s.Reason)
	if s.Sentry {
		result = result.WithSentry()
	}
	return result
}

func (s ErrorStopChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	return subroutine.OK()
}

func (s ErrorStopChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "ErrorStopChainSubroutine"
}

func (s ErrorStopChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return nil
}

type TrackingChainSubroutine struct {
	Name           string
	ProcessCalled  bool
	FinalizeCalled bool
	ReturnResult   subroutine.Result
}

func (s *TrackingChainSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	s.ProcessCalled = true
	return s.ReturnResult
}

func (s *TrackingChainSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	s.FinalizeCalled = true
	return s.ReturnResult
}

func (s *TrackingChainSubroutine) GetName() string {
	if s.Name != "" {
		return s.Name
	}
	return "TrackingChainSubroutine"
}

func (s *TrackingChainSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{ChainSubroutineFinalizer}
}
