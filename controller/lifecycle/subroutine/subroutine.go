package subroutine

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/errors"
)

type Subroutine interface {
	Process(ctx context.Context, instance runtimeobject.RuntimeObject) (ctrl.Result, errors.OperatorError)
	Finalize(ctx context.Context, instance runtimeobject.RuntimeObject) (ctrl.Result, errors.OperatorError)
	GetName() string
	Finalizers(instance runtimeobject.RuntimeObject) []string
	// ShouldStopChain returns true if the reconciliation chain should stop after this subroutine
	// this allows subroutines to signal that subsequent subroutines should not run (e.g. domain filtering)
	ShouldStopChain() bool
}
