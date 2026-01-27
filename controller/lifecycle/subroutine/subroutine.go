package subroutine

import (
	"context"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
)

type Subroutine interface {
	Process(ctx context.Context, instance runtimeobject.RuntimeObject) Result
	Finalize(ctx context.Context, instance runtimeobject.RuntimeObject) Result
	GetName() string
	Finalizers(instance runtimeobject.RuntimeObject) []string
}
