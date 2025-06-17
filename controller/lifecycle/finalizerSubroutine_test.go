package lifecycle

import (
	"context"
	"time"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/platform-mesh/golang-commons/controller/testSupport"
	"github.com/platform-mesh/golang-commons/errors"
)

const subroutineFinalizer = "finalizer"

type finalizerSubroutine struct {
	client       client.Client
	err          error
	requeueAfter time.Duration
}

func (c finalizerSubroutine) Process(_ context.Context, runtimeObj RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	instance := runtimeObj.(*testSupport.TestApiObject)
	instance.Status.Some = "other string"
	return controllerruntime.Result{}, nil
}

func (c finalizerSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if c.err != nil {
		return controllerruntime.Result{}, errors.NewOperatorError(c.err, true, true)
	}
	if c.requeueAfter > 0 {
		return controllerruntime.Result{RequeueAfter: c.requeueAfter}, nil
	}

	return controllerruntime.Result{}, nil
}

func (c finalizerSubroutine) GetName() string {
	return "changeStatus"
}

func (c finalizerSubroutine) Finalizers() []string {
	return []string{
		subroutineFinalizer,
	}
}
