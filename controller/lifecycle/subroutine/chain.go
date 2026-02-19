package subroutine

import (
	"context"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
)

// ChainSubroutine is an interface for subroutines with chain control capabilities
// This interface supports additional outcomes like StopChain, ErrorContinue, and Skipped
// For simple subroutines where errors should stop reconciliation, use Subroutine instead
type ChainSubroutine interface {
	BaseSubroutine
	Process(ctx context.Context, instance runtimeobject.RuntimeObject) Result
	Finalize(ctx context.Context, instance runtimeobject.RuntimeObject) Result
}
