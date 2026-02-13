package subroutine

import (
	"context"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ChainSubroutine is an interface for subroutines with chain control capabilities
// This interface supports additional outcomes like StopChain and Skipped
// NOTE: use ChainAdapter to wrap implementations for use with the lifecycle
type ChainSubroutine interface {
	Process(ctx context.Context, instance runtimeobject.RuntimeObject) Result
	Finalize(ctx context.Context, instance runtimeobject.RuntimeObject) Result
	GetName() string
	Finalizers(instance runtimeobject.RuntimeObject) []string
}

// ChainAdapter wraps a ChainSubroutine to implement the Subroutine interface
// This allows chain subroutines to be used with the existing lifecycle infrastructure
// The lifecycle detects the underlying ChainSubroutine implementation via Unwrap() and uses
// the chain methods directly when available
type ChainAdapter struct {
	chain ChainSubroutine
}

// NewChainAdapter is a factory method that creates an adapter that wraps a ChainSubroutine implementation
func NewChainAdapter(chain ChainSubroutine) *ChainAdapter {
	return &ChainAdapter{chain: chain}
}

// Process implements Subroutine by delegating to the ChainSubroutine implementation
// This is only called if the lifecycle doesn't detect the ChainSubroutine interface
func (a *ChainAdapter) Process(
	ctx context.Context,
	instance runtimeobject.RuntimeObject,
) (ctrl.Result, errors.OperatorError) {
	result := a.chain.Process(ctx, instance)
	return result.ToSubroutine()
}

// Finalize implements Subroutine by delegating to the ChainSubroutine implementation
// This is only called if the lifecycle doesn't detect the ChainSubroutine interface
func (a *ChainAdapter) Finalize(
	ctx context.Context,
	instance runtimeobject.RuntimeObject,
) (ctrl.Result, errors.OperatorError) {
	result := a.chain.Finalize(ctx, instance)
	return result.ToSubroutine()
}

func (a *ChainAdapter) GetName() string {
	return a.chain.GetName()
}

func (a *ChainAdapter) Finalizers(instance runtimeobject.RuntimeObject) []string {
	return a.chain.Finalizers(instance)
}

// Unwrap returns the underlying ChainSubroutine implementation
// This is used by the lifecycle to detect chain subroutines and use them directly
func (a *ChainAdapter) Unwrap() ChainSubroutine {
	return a.chain
}
