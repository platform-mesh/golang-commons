package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"

	pmtesting "github.com/platform-mesh/golang-commons/controller/testSupport"
	"github.com/platform-mesh/golang-commons/logger"
)

func TestNewBuilder_Defaults(t *testing.T) {
	log := &logger.Logger{}
	b := NewBuilder("op", "ctrl", nil, log)
	if b.operatorName != "op" {
		t.Errorf("expected operatorName 'op', got %s", b.operatorName)
	}
	if b.controllerName != "ctrl" {
		t.Errorf("expected controllerName 'ctrl', got %s", b.controllerName)
	}
	if b.withConditionManagement {
		t.Error("expected withConditionManagement to be false")
	}
	if b.withSpreadingReconciles {
		t.Error("expected withSpreadingReconciles to be false")
	}
	if b.withReadOnly {
		t.Error("expected withReadOnly to be false")
	}
	if b.log != log {
		t.Error("expected log to be set")
	}
}

func TestBuilder_WithConditionManagement(t *testing.T) {
	b := NewBuilder("op", "ctrl", nil, &logger.Logger{})
	b.WithConditionManagement()
	if !b.withConditionManagement {
		t.Error("WithConditionManagement should set withConditionManagement to true")
	}
}

func TestBuilder_WithSpreadingReconciles(t *testing.T) {
	b := NewBuilder("op", "ctrl", nil, &logger.Logger{})
	b.WithSpreadingReconciles()
	if !b.withSpreadingReconciles {
		t.Error("WithSpreadingReconciles should set withSpreadingReconciles to true")
	}
}

func TestBuilder_WithReadOnly(t *testing.T) {
	b := NewBuilder("op", "ctrl", nil, &logger.Logger{})
	b.WithReadOnly()
	if !b.withReadOnly {
		t.Error("WithReadOnly should set withReadOnly to true")
	}
}

func TestControllerRuntimeBuilder(t *testing.T) {
	t.Run("Minimal setup", func(t *testing.T) {
		b := NewBuilder("op", "ctrl", nil, &logger.Logger{})
		fakeClient := pmtesting.CreateFakeClient(t, &pmtesting.TestApiObject{})
		lm := b.BuildControllerRuntime(fakeClient)
		assert.NotNil(t, lm)
	})
	t.Run("All Options", func(t *testing.T) {
		b := NewBuilder("op", "ctrl", nil, &logger.Logger{}).WithConditionManagement().WithSpreadingReconciles()
		fakeClient := pmtesting.CreateFakeClient(t, &pmtesting.TestApiObject{})
		lm := b.BuildControllerRuntime(fakeClient)
		assert.NotNil(t, lm)
	})
	t.Run("ReadOnly", func(t *testing.T) {
		b := NewBuilder("op", "ctrl", nil, &logger.Logger{}).WithReadOnly()
		fakeClient := pmtesting.CreateFakeClient(t, &pmtesting.TestApiObject{})
		lm := b.BuildControllerRuntime(fakeClient)
		assert.NotNil(t, lm)
	})
}

func TestMulticontrollerRuntimeBuilder(t *testing.T) {
	t.Run("Minimal setup", func(t *testing.T) {
		b := NewBuilder("op", "ctrl", nil, &logger.Logger{})
		cfg := &rest.Config{}
		provider := pmtesting.NewFakeProvider(cfg)
		mgr, err := mcmanager.New(cfg, provider, mcmanager.Options{})
		assert.NoError(t, err)
		lm := b.BuildMultiCluster(mgr)
		assert.NotNil(t, lm)
	})
	t.Run("All Options", func(t *testing.T) {
		b := NewBuilder("op", "ctrl", nil, &logger.Logger{}).WithConditionManagement().WithSpreadingReconciles()
		cfg := &rest.Config{}
		provider := pmtesting.NewFakeProvider(cfg)
		mgr, err := mcmanager.New(cfg, provider, mcmanager.Options{})
		assert.NoError(t, err)
		lm := b.BuildMultiCluster(mgr)
		assert.NotNil(t, lm)
	})
	t.Run("ReadOnly", func(t *testing.T) {
		b := NewBuilder("op", "ctrl", nil, &logger.Logger{}).WithReadOnly()
		cfg := &rest.Config{}
		provider := pmtesting.NewFakeProvider(cfg)
		mgr, err := mcmanager.New(cfg, provider, mcmanager.Options{})
		assert.NoError(t, err)
		lm := b.BuildMultiCluster(mgr)
		assert.NotNil(t, lm)
	})
}
