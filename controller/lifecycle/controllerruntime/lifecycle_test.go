package controllerruntime

import (
	"context"
	goerrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	pmtesting "github.com/platform-mesh/golang-commons/controller/testSupport"
	operrors "github.com/platform-mesh/golang-commons/errors"
	"github.com/platform-mesh/golang-commons/logger/testlogger"
)

func TestLifecycle(t *testing.T) {
	namespace := "bar"
	name := "foo"
	testApiObject := &pmtesting.TestApiObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	t.Run("Test Lifecycle setupWithManager /w spread and expecting no error", func(t *testing.T) {
		// Arrange
		instance := &pmtesting.ImplementingSpreadReconciles{}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		m, err := manager.New(&rest.Config{}, manager.Options{Scheme: fakeClient.Scheme()})
		assert.NoError(t, err)

		lm, log := createLifecycleManager([]subroutine.Subroutine{}, fakeClient)
		lm = lm.WithSpreadingReconciles()
		tr := &testReconciler{lifecycleManager: lm}

		// Act
		err = lm.SetupWithManager(m, 0, "testReconciler3", instance, "test", tr, log.Logger)

		// Assert
		assert.NoError(t, err)
	})
	t.Run("Test Lifecycle setupWithManager /w spread and expecting a error", func(t *testing.T) {
		// Arrange
		instance := &pmtesting.NotImplementingSpreadReconciles{}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		m, err := manager.New(&rest.Config{}, manager.Options{Scheme: fakeClient.Scheme()})
		assert.NoError(t, err)

		lm, log := createLifecycleManager([]subroutine.Subroutine{}, fakeClient)
		lm = lm.WithSpreadingReconciles()
		tr := &testReconciler{lifecycleManager: lm}

		// Act
		err = lm.SetupWithManager(m, 0, "testReconciler", instance, "test", tr, log.Logger)

		// Assert
		assert.Error(t, err)
	})
	t.Run("Test Lifecycle setupWithManager /w spread and expecting a error (invalid config)", func(t *testing.T) {
		// Arrange
		instance := &pmtesting.ImplementingSpreadReconciles{}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		m, err := manager.New(&rest.Config{}, manager.Options{Scheme: fakeClient.Scheme()})
		assert.NoError(t, err)

		lm, log := createLifecycleManager([]subroutine.Subroutine{}, fakeClient)
		lm.WithSpreadingReconciles().WithReadOnly()
		tr := &testReconciler{lifecycleManager: lm}

		// Act
		err = lm.SetupWithManager(m, 0, "testReconciler", instance, "test", tr, log.Logger)

		// Assert
		assert.Error(t, err)
	})
	errorMessage := "oh nose"
	t.Run("Prepare Context", func(t *testing.T) {
		t.Run("Sets a context that can be used in the subroutine", func(t *testing.T) {
			// Arrange
			ctx := context.Background()

			fakeClient := pmtesting.CreateFakeClient(t, testApiObject)

			lm, _ := createLifecycleManager([]subroutine.Subroutine{pmtesting.ContextValueSubroutine{}}, fakeClient)
			lm = lm.WithPrepareContextFunc(func(ctx context.Context, instance runtimeobject.RuntimeObject) (context.Context, operrors.OperatorError) {
				return context.WithValue(ctx, pmtesting.ContextValueKey, "valueFromContext"), nil
			})
			tr := &testReconciler{lifecycleManager: lm}
			result, err := tr.Reconcile(ctx, controllerruntime.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}})

			// Then
			assert.NotNil(t, ctx)
			assert.NotNil(t, result)
			assert.NoError(t, err)

			err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, testApiObject)
			assert.NoError(t, err)
			assert.Equal(t, "valueFromContext", testApiObject.Status.Some)
		})

		t.Run("Handles the errors correctly", func(t *testing.T) {
			// Arrange
			ctx := context.Background()

			fakeClient := pmtesting.CreateFakeClient(t, testApiObject)

			lm, _ := createLifecycleManager([]subroutine.Subroutine{pmtesting.ContextValueSubroutine{}}, fakeClient)
			lm = lm.WithPrepareContextFunc(func(ctx context.Context, instance runtimeobject.RuntimeObject) (context.Context, operrors.OperatorError) {
				return nil, operrors.NewOperatorError(goerrors.New(errorMessage), true, false)
			})
			tr := &testReconciler{lifecycleManager: lm}
			result, err := tr.Reconcile(ctx, controllerruntime.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}})

			// Then
			assert.NotNil(t, ctx)
			assert.NotNil(t, result)
			assert.Error(t, err)
		})
	})
	t.Run("WthConditionManagement", func(t *testing.T) {
		// Given
		fakeClient := pmtesting.CreateFakeClient(t, &pmtesting.TestApiObject{})
		_, log := createLifecycleManager([]subroutine.Subroutine{}, fakeClient)

		// When
		l := NewLifecycleManager(log.Logger, "test-operator", "test-controller", fakeClient, []subroutine.Subroutine{}).WithConditionManagement()

		// Then
		assert.True(t, true, l.ConditionsManager() != nil)
	})
	t.Run("WithReadOnly", func(t *testing.T) {
		// Given
		fakeClient := pmtesting.CreateFakeClient(t, &pmtesting.TestApiObject{})
		_, log := createLifecycleManager([]subroutine.Subroutine{}, fakeClient)

		// When
		l := NewLifecycleManager(log.Logger, "test-operator", "test-controller", fakeClient, []subroutine.Subroutine{}).WithReadOnly()

		// Then
		assert.True(t, true, l.ConditionsManager() != nil)
	})

}

type testReconciler struct {
	lifecycleManager *LifecycleManager
}

func (r *testReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	return r.lifecycleManager.Reconcile(ctx, req, &pmtesting.TestApiObject{})
}

func createLifecycleManager(subroutines []subroutine.Subroutine, c client.Client) (*LifecycleManager, *testlogger.TestLogger) {
	log := testlogger.New()
	mgr := NewLifecycleManager(log.Logger, "test-operator", "test-controller", c, subroutines)
	return mgr, log
}
