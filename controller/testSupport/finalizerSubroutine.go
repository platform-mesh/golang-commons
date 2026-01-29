package testSupport

import (
	"context"
	"time"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
)

const SubroutineFinalizer = "finalizer"

type FinalizerSubroutine struct {
	Client       client.Client
	Err          error
	RequeueAfter time.Duration
}

func (c FinalizerSubroutine) Process(_ context.Context, runtimeObj runtimeobject.RuntimeObject) subroutine.Result {
	instance := runtimeObj.(*TestApiObject)
	instance.Status.Some = "other string"
	return subroutine.OK()
}

func (c FinalizerSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) subroutine.Result {
	if c.Err != nil {
		return subroutine.Retry(c.Err)
	}
	if c.RequeueAfter > 0 {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter})
	}

	return subroutine.OK()
}

func (c FinalizerSubroutine) GetName() string {
	return "changeStatus"
}

func (c FinalizerSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{
		SubroutineFinalizer,
	}
}
