package lifecycle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	"github.com/openmfp/golang-commons/logger/testlogger"
)

func TestLifecycle(t *testing.T) {
	namespace := "bar"
	name := "foo"
	request := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}
	ctx := context.Background()

	t.Run("Lifecycle without changing status", func(t *testing.T) {
		// Arrange
		instance := &testApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: TestStatus{Some: "string"},
		}
		fakeClient := createFakeClient(t, instance)

		manager, logger := createLifecycleManager([]Subroutine{}, fakeClient)

		// Act
		result, err := manager.Reconcile(ctx, request, instance)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		logMessages, err := logger.GetLogMessages()
		assert.NoError(t, err)
		assert.Equal(t, len(logMessages), 3)
		assert.Equal(t, logMessages[0].Message, "start reconcile")
		assert.Equal(t, logMessages[1].Message, "skipping status update, since they are equal")
		assert.Equal(t, logMessages[2].Message, "end reconcile")
	})

	t.Run("Lifecycle with changing status", func(t *testing.T) {
		// Arrange
		instance := &testApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: TestStatus{Some: "string"},
		}

		fakeClient := createFakeClient(t, instance)

		manager, logger := createLifecycleManager([]Subroutine{
			changeStatusSubroutine{
				client: fakeClient,
			},
		}, fakeClient)

		// Act
		result, err := manager.Reconcile(ctx, request, instance)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		logMessages, err := logger.GetLogMessages()
		assert.NoError(t, err)
		assert.Equal(t, len(logMessages), 5)
		assert.Equal(t, logMessages[0].Message, "start reconcile")
		assert.Equal(t, logMessages[1].Message, "start subroutine")
		assert.Equal(t, logMessages[2].Message, "end subroutine")
		assert.Equal(t, logMessages[3].Message, "updating resource status")
		assert.Equal(t, logMessages[4].Message, "end reconcile")

		serverObject := &testApiObject{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, serverObject)
		assert.NoError(t, err)
		assert.Equal(t, serverObject.Status.Some, "other string")
	})

}

func createLifecycleManager(subroutines []Subroutine, c client.Client) (*LifecycleManager, *testlogger.TestLogger) {
	logger := testlogger.New()

	manager := NewLifecycleManager(logger.Logger, "test-operator", "test-controller", c, subroutines)
	return manager, logger
}
func createFakeClient(t *testing.T, objects ...runtime.Object) client.WithWatch {
	builder := fake.NewClientBuilder()
	s := runtime.NewScheme()
	sBuilder := scheme.Builder{GroupVersion: schema.GroupVersion{Group: "test.openmfp.com", Version: "v1alpha1"}}
	object := &testApiObject{}
	sBuilder.Register(object)
	err := sBuilder.AddToScheme(s)
	assert.NoError(t, err)
	builder.WithScheme(s)
	builder.WithStatusSubresource(object)
	builder.WithRuntimeObjects(objects...)
	return builder.Build()
}
