package lifecycle

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	pmtesting "github.com/platform-mesh/golang-commons/controller/testSupport"
	"github.com/platform-mesh/golang-commons/logger"
)

func TestChainSubroutine(t *testing.T) {
	namespace := "test-ns"
	name := "test-obj"
	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}
	ctx := context.Background()
	logcfg := logger.DefaultConfig()
	logcfg.NoJSON = true
	log, err := logger.New(logcfg)
	require.NoError(t, err)

	t.Run("ChainSubroutine with OK result", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.OKChainSubroutine{}.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Zero(t, result.RequeueAfter)
	})

	t.Run("ChainSubroutine changes status", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: pmtesting.TestStatus{Some: "original"},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.ChangeStatusChainSubroutine{Client: fakeClient}.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		serverObject := &pmtesting.TestApiObject{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, serverObject)
		assert.NoError(t, err)
		assert.Equal(t, "changed by chain", serverObject.Status.Some)
	})

	t.Run("ChainSubroutine StopChain stops subsequent subroutines", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "TrackingAfterStop",
			ReturnResult: subroutine.OK(),
		}

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.StopChainSubroutine{Name: "StopFirst", Reason: "test stop"}.Adapted(),
				tracker.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, tracker.ProcessCalled, "Subroutine after StopChain should not be called")
	})

	t.Run("ChainSubroutine StopChain with requeue", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.StopChainSubroutine{
					Name:         "StopWithRequeue",
					Reason:       "waiting",
					RequeueAfter: 30 * time.Second,
				}.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, result.RequeueAfter)
	})

	t.Run("ChainSubroutine ErrorRetry returns error and stops chain", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		retryErr := fmt.Errorf("retry this error")
		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "TrackingAfterRetry",
			ReturnResult: subroutine.OK(),
		}

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.RetryChainSubroutine{Name: "RetryFirst", Err: retryErr, Sentry: true}.Adapted(),
				tracker.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.Error(t, err)
		assert.Equal(t, retryErr, err)
		assert.NotNil(t, result)
		assert.False(t, tracker.ProcessCalled, "Subroutine after ErrorRetry should not be called")
	})

	t.Run("ChainSubroutine ErrorContinue continues to next subroutine", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		failErr := fmt.Errorf("non-retryable error")
		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "TrackingAfterFail",
			ReturnResult: subroutine.OK(),
		}

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.FailChainSubroutine{Name: "FailFirst", Err: failErr, Sentry: true}.Adapted(),
				tracker.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err) // ErrorContinue must not return error from reconcile
		assert.NotNil(t, result)
		assert.True(t, tracker.ProcessCalled, "Subroutine after ErrorContinue should be called")
	})

	t.Run("ChainSubroutine mixed with v1 subroutines", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: pmtesting.TestStatus{Some: "original"},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "TrackingV2",
			ReturnResult: subroutine.OK(),
		}

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.ChangeStatusSubroutine{Client: fakeClient}, // standard subroutine
				tracker.Adapted(), // chainsubroutine
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, tracker.ProcessCalled, "V2 subroutine should be called after v1")
	})

	t.Run("ChainSubroutine finalization removes finalizers", func(t *testing.T) {
		now := &metav1.Time{Time: time.Now()}
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        []string{pmtesting.ChainSubroutineFinalizer},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.FinalizerChainSubroutine{Client: fakeClient}.Adapted(),
			},
		}

		_, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(instance.Finalizers), "Finalizer should be removed")
	})

	t.Run("ChainSubroutine finalization with error retries", func(t *testing.T) {
		now := &metav1.Time{Time: time.Now()}
		finalizeErr := fmt.Errorf("finalize error")
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        []string{pmtesting.ChainSubroutineFinalizer},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.FinalizerChainSubroutine{Client: fakeClient, Err: finalizeErr}.Adapted(),
			},
		}

		_, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.Error(t, err)
		assert.Equal(t, finalizeErr, err)
		assert.Equal(t, 1, len(instance.Finalizers), "Finalizer should not be removed on error")
	})

	t.Run("ChainSubroutine finalization with requeue keeps finalizer", func(t *testing.T) {
		now := &metav1.Time{Time: time.Now()}
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        []string{pmtesting.ChainSubroutineFinalizer},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.FinalizerChainSubroutine{Client: fakeClient, RequeueAfter: 5 * time.Second}.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, result.RequeueAfter)
		assert.Equal(t, 1, len(instance.Finalizers), "Finalizer should not be removed when requeue requested")
	})

	t.Run("ChainSubroutine skips finalization when no finalizer present", func(t *testing.T) {
		now := &metav1.Time{Time: time.Now()}
		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "TrackerNoFinalizer",
			ReturnResult: subroutine.OK(),
		}
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        []string{"other-finalizer"},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				tracker.Adapted(),
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, tracker.FinalizeCalled, "Finalize should not be called when no matching finalizer")
	})

	t.Run("ChainSubroutine with conditions manager - OK sets condition", func(t *testing.T) {
		instance := &pmtesting.ImplementConditions{
			TestApiObject: pmtesting.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: pmtesting.TestStatus{},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.OKChainSubroutine{}.Adapted(),
			},
		}
		mgr.WithConditionManagement()

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(instance.Status.Conditions), 1)
	})

	t.Run("ChainSubroutine with conditions manager - StopChain sets condition", func(t *testing.T) {

		instance := &pmtesting.ImplementConditions{
			TestApiObject: pmtesting.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: pmtesting.TestStatus{},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.StopChainSubroutine{Reason: "waiting for dependency"}.Adapted(),
			},
		}
		mgr.WithConditionManagement()

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(instance.Status.Conditions), 1)
	})

	t.Run("ChainSubroutine ErrorContinue marks resource as not ready", func(t *testing.T) {
		instance := &pmtesting.ImplementConditionsAndSpreadReconciles{
			TestApiObject: pmtesting.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: pmtesting.TestStatus{
					Some:               "string",
					ObservedGeneration: 0,
				},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger:          log,
			ShouldReconcile: true,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.FailChainSubroutine{Err: fmt.Errorf("non-retryable"), Sentry: false}.Adapted(),
			},
		}
		mgr.WithSpreadingReconciles()
		mgr.WithConditionManagement()

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// With non retryable error, observedGeneration should still be updated
		assert.Equal(t, int64(1), instance.Status.ObservedGeneration)
	})

	t.Run("ChainSubroutine multiple subroutines - minimum requeue time is used", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		// 1st subroutine with 30 second requeue
		sub1 := &pmtesting.TrackingChainSubroutine{
			Name:         "Sub1",
			ReturnResult: subroutine.OKWithRequeue(ctrl.Result{RequeueAfter: 30 * time.Second}),
		}
		// 2nd subroutine with 10 second requeue (should be used)
		sub2 := &pmtesting.TrackingChainSubroutine{
			Name:         "Sub2",
			ReturnResult: subroutine.OKWithRequeue(ctrl.Result{RequeueAfter: 10 * time.Second}),
		}

		mgr := &pmtesting.TestLifecycleManager{
			Logger:         log,
			SubroutinesArr: []subroutine.Subroutine{sub1.Adapted(), sub2.Adapted()},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, result.RequeueAfter)
		assert.True(t, sub1.ProcessCalled)
		assert.True(t, sub2.ProcessCalled)
	})

	t.Run("ChainSubroutine adds finalizers when processing", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger:         log,
			SubroutinesArr: []subroutine.Subroutine{pmtesting.FinalizerChainSubroutine{Client: fakeClient}.Adapted()},
		}

		_, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.Contains(t, instance.Finalizers, pmtesting.ChainSubroutineFinalizer)
	})

	t.Run("ChainSubroutine RetryNoSentry does not send to sentry", func(t *testing.T) {
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.RetryChainSubroutine{Err: fmt.Errorf("expected error"), Sentry: false}.Adapted(),
			},
		}

		_, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.Error(t, err)
	})

	t.Run("ChainSubroutine FailWithReason includes reason", func(t *testing.T) {
		instance := &pmtesting.ImplementConditions{
			TestApiObject: pmtesting.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: pmtesting.TestStatus{},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.FailChainSubroutine{Err: fmt.Errorf("config error"), Reason: "bad config"}.Adapted(),
			},
		}
		mgr.WithConditionManagement()

		_, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(instance.Status.Conditions), 1)
	})

	t.Run("ChainSubroutine v1 after v2 both execute", func(t *testing.T) {

		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: pmtesting.TestStatus{Some: "original"},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "TrackingV2First",
			ReturnResult: subroutine.OK(),
		}

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				tracker.Adapted(),
				pmtesting.ChangeStatusSubroutine{Client: fakeClient},
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, tracker.ProcessCalled, "V2 subroutine should be called")

		serverObject := &pmtesting.TestApiObject{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, serverObject)
		assert.NoError(t, err)
		assert.Equal(t, "other string", serverObject.Status.Some)
	})

	t.Run("ChainSubroutine StopChain prevents v1 subroutine from running", func(t *testing.T) {

		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: pmtesting.TestStatus{Some: "original"},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{
			Logger: log,
			SubroutinesArr: []subroutine.Subroutine{
				pmtesting.StopChainSubroutine{Reason: "stop"}.Adapted(),
				pmtesting.ChangeStatusSubroutine{Client: fakeClient}, // standard subroutine msut not be called
			},
		}

		result, err := Reconcile(ctx, request.NamespacedName, instance, fakeClient, mgr)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// status must not have changed because standard subroutine was not called
		serverObject := &pmtesting.TestApiObject{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, serverObject)
		assert.NoError(t, err)
		assert.Equal(t, "original", serverObject.Status.Some)
	})
}

func TestReconcileChainSubroutine(t *testing.T) {
	logcfg := logger.DefaultConfig()
	logcfg.NoJSON = true
	log, err := logger.New(logcfg)
	require.NoError(t, err)

	t.Run("reconcileChainSubroutine returns Skip when no finalizer during deletion", func(t *testing.T) {
		now := &metav1.Time{Time: time.Now()}
		instance := &pmtesting.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test",
				Namespace:         "test-ns",
				DeletionTimestamp: now,
				Finalizers:        []string{"other-finalizer"},
			},
		}
		fakeClient := pmtesting.CreateFakeClient(t, instance)

		mgr := &pmtesting.TestLifecycleManager{Logger: log}
		sub := pmtesting.OKChainSubroutine{}

		result := reconcileChainSubroutine(
			context.Background(),
			instance,
			sub,
			fakeClient,
			mgr,
			log,
			true,
			nil,
		)

		assert.Equal(t, subroutine.Skipped, result.Outcome)
		assert.Equal(t, "no finalizer", result.Reason)
	})
}

func TestChainAdapter(t *testing.T) {
	t.Run("ChainAdapter Process delegates to V2", func(t *testing.T) {
		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "Tracker",
			ReturnResult: subroutine.OK(),
		}
		adapter := subroutine.NewChainAdapter(tracker)

		result, err := adapter.Process(context.Background(), &pmtesting.TestApiObject{})

		assert.Nil(t, err)
		assert.Equal(t, ctrl.Result{}, result)
		assert.True(t, tracker.ProcessCalled)
	})

	t.Run("ChainAdapter Finalize delegates to V2", func(t *testing.T) {
		tracker := &pmtesting.TrackingChainSubroutine{
			Name:         "Tracker",
			ReturnResult: subroutine.OK(),
		}
		adapter := subroutine.NewChainAdapter(tracker)

		result, err := adapter.Finalize(context.Background(), &pmtesting.TestApiObject{})

		assert.Nil(t, err)
		assert.Equal(t, ctrl.Result{}, result)
		assert.True(t, tracker.FinalizeCalled)
	})

	t.Run("ChainAdapter converts error results correctly", func(t *testing.T) {
		testErr := fmt.Errorf("test error")
		sub := pmtesting.RetryChainSubroutine{Err: testErr, Sentry: true}
		adapter := subroutine.NewChainAdapter(sub)

		result, err := adapter.Process(context.Background(), &pmtesting.TestApiObject{})

		assert.NotNil(t, err)
		assert.Equal(t, testErr, err.Err())
		assert.True(t, err.Retry())
		assert.True(t, err.Sentry())
		assert.Equal(t, ctrl.Result{}, result)
	})

	t.Run("ChainAdapter GetName delegates to V2", func(t *testing.T) {
		sub := pmtesting.OKChainSubroutine{Name: "TestName"}
		adapter := subroutine.NewChainAdapter(sub)
		assert.Equal(t, "TestName", adapter.GetName())
	})

	t.Run("ChainAdapter Finalizers delegates to V2", func(t *testing.T) {
		sub := pmtesting.FinalizerChainSubroutine{}
		adapter := subroutine.NewChainAdapter(sub)

		assert.Equal(t, []string{pmtesting.ChainSubroutineFinalizer}, adapter.Finalizers(&pmtesting.TestApiObject{}))
	})

	t.Run("ChainAdapter Unwrap returns underlying V2", func(t *testing.T) {
		sub := pmtesting.OKChainSubroutine{Name: "TestName"}
		adapter := subroutine.NewChainAdapter(sub)

		unwrapped := adapter.Unwrap()

		assert.Equal(t, sub, unwrapped)
	})
}
