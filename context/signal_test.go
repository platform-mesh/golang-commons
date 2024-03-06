package context

import (
	"context"
	"errors"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotifyContextSIGINT(t *testing.T) {
	ctx, _ := NotifyShutdownContext(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		wg.Done()
	}()

	wg.Wait()
	<-ctx.Done()
	err := ctx.Err()
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))

	cause := context.Cause(ctx)
	assert.NotNil(t, cause)
	assert.True(t, errors.Is(cause, ShutdownError))
}

func TestNotifyContextSIGTERM(t *testing.T) {
	ctx, _ := NotifyShutdownContext(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		wg.Done()
	}()

	wg.Wait()
	<-ctx.Done()
	err := ctx.Err()
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))

	cause := context.Cause(ctx)
	assert.NotNil(t, cause)
	assert.True(t, errors.Is(cause, ShutdownError))
}
