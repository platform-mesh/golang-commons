package lifecycle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/mocks"
	"github.com/platform-mesh/golang-commons/controller/testSupport"
	pmtesting "github.com/platform-mesh/golang-commons/controller/testSupport"
	"github.com/platform-mesh/golang-commons/logger"
)

func TestLifecycle(t *testing.T) {
	namespace := "bar"
	name := "foo"
	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}
	logcfg := logger.DefaultConfig()
	logcfg.NoJSON = true
	log, err := logger.New(logcfg)
	assert.NoError(t, err)
	//testApiObject := &testSupport.TestApiObject{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      name,
	//		Namespace: namespace,
	//	},
	//}
	ctx := context.Background()

	t.Run("Lifecycle with a not found object", func(t *testing.T) {
		// Arrange
		fakeClient := testSupport.CreateFakeClient(t, &testSupport.TestApiObject{})

		mgr := pmtesting.TestLifecycleManager{Logger: log}

		// Act
		result, err := Reconcile(ctx, request, &testSupport.TestApiObject{}, fakeClient, mgr)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NoError(t, err)
	})

}

func TestUpdateStatus(t *testing.T) {
	clientMock := new(mocks.Client)
	subresourceClient := new(mocks.SubResourceClient)

	logcfg := logger.DefaultConfig()
	logcfg.NoJSON = true
	log, err := logger.New(logcfg)
	assert.NoError(t, err)

	t.Run("Test UpdateStatus with no changes", func(t *testing.T) {
		original := &pmtesting.ImplementingSpreadReconciles{
			TestApiObject: testSupport.TestApiObject{
				Status: testSupport.TestStatus{
					Some: "string",
				},
			}}

		// When
		err := updateStatus(context.Background(), clientMock, original, original, log, true, nil)

		// Then
		assert.NoError(t, err)
	})

	t.Run("Test UpdateStatus with update error", func(t *testing.T) {
		original := &pmtesting.ImplementingSpreadReconciles{
			TestApiObject: testSupport.TestApiObject{
				Status: testSupport.TestStatus{
					Some: "string",
				},
			}}
		current := &pmtesting.ImplementingSpreadReconciles{
			TestApiObject: testSupport.TestApiObject{
				Status: testSupport.TestStatus{
					Some: "string1",
				},
			}}

		clientMock.EXPECT().Status().Return(subresourceClient)
		subresourceClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
			Return(errors.NewBadRequest("internal error"))

		// When
		err := updateStatus(context.Background(), clientMock, original, current, log, true, nil)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "internal error", err.Error())
	})

	t.Run("Test UpdateStatus with no status object (original)", func(t *testing.T) {
		original := &testSupport.TestNoStatusApiObject{}
		current := &pmtesting.ImplementConditions{}
		// When
		err := updateStatus(context.Background(), clientMock, original, current, log, true, nil)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "status field not found in current object", err.Error())
	})
	t.Run("Test UpdateStatus with no status object (current)", func(t *testing.T) {
		original := &pmtesting.ImplementConditions{}
		current := &testSupport.TestNoStatusApiObject{}
		// When
		err := updateStatus(context.Background(), clientMock, original, current, log, true, nil)

		// Then
		assert.Error(t, err)
		assert.Equal(t, "status field not found in current object", err.Error())
	})
}
