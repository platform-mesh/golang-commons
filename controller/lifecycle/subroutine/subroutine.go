package subroutine

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/errors"
)

type BaseSubroutine interface {
	GetName() string
	Finalizers(instance runtimeobject.RuntimeObject) []string
}

type Subroutine interface {
	BaseSubroutine
	Process(ctx context.Context, instance runtimeobject.RuntimeObject) (ctrl.Result, errors.OperatorError)
	Finalize(ctx context.Context, instance runtimeobject.RuntimeObject) (ctrl.Result, errors.OperatorError)
}
