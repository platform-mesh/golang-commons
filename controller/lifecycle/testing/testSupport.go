package testing

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/platform-mesh/golang-commons/context/keys"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/platform-mesh/golang-commons/controller/testSupport"
	"github.com/platform-mesh/golang-commons/errors"
)

const FailureScenarioSubroutineFinalizer = "failuresubroutine"
const ChangeStatusSubroutineFinalizer = "changestatus"

type ImplementConditions struct {
	testSupport.TestApiObject `json:",inline"`
}

func (m *ImplementConditions) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *ImplementConditions) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

type ImplementingSpreadReconciles struct {
	testSupport.TestApiObject `json:",inline"`
}

func (m *ImplementingSpreadReconciles) GetGeneration() int64 {
	return m.Generation
}

func (m *ImplementingSpreadReconciles) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}

func (m *ImplementingSpreadReconciles) SetObservedGeneration(g int64) {
	m.Status.ObservedGeneration = g
}

func (m *ImplementingSpreadReconciles) GetNextReconcileTime() metav1.Time {
	return m.Status.NextReconcileTime
}

func (m *ImplementingSpreadReconciles) SetNextReconcileTime(time metav1.Time) {
	m.Status.NextReconcileTime = time
}

type NotImplementingSpreadReconciles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status testSupport.TestStatus `json:"status,omitempty"`
}

func (m *NotImplementingSpreadReconciles) DeepCopyObject() runtime.Object {
	if c := m.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (m *NotImplementingSpreadReconciles) DeepCopy() *NotImplementingSpreadReconciles {
	if m == nil {
		return nil
	}
	out := new(NotImplementingSpreadReconciles)
	m.DeepCopyInto(out)
	return out
}
func (m *NotImplementingSpreadReconciles) DeepCopyInto(out *NotImplementingSpreadReconciles) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = m.Status
}

type ChangeStatusSubroutine struct {
	Client client.Client
}

func (c ChangeStatusSubroutine) Process(_ context.Context, runtimeObj runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if instance, ok := runtimeObj.(*testSupport.TestApiObject); ok {
		instance.Status.Some = "other string"
	}
	if instance, ok := runtimeObj.(*ImplementingSpreadReconciles); ok {
		instance.Status.Some = "other string"
	}

	if instance, ok := runtimeObj.(*ImplementConditions); ok {
		instance.Status.Some = "other string"
	}
	return controllerruntime.Result{}, nil
}

func (c ChangeStatusSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (c ChangeStatusSubroutine) GetName() string {
	return "changeStatus"
}

func (c ChangeStatusSubroutine) Finalizers() []string {
	return []string{"changestatus"}
}

type AddConditionSubroutine struct {
	Ready metav1.ConditionStatus
}

func (c AddConditionSubroutine) Process(_ context.Context, runtimeObj runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if instance, ok := runtimeObj.(*ImplementConditions); ok {
		instance.Status.Some = "other string"
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:    "test",
			Status:  c.Ready,
			Reason:  "test",
			Message: "test",
		})
	}

	return controllerruntime.Result{}, nil
}

func (c AddConditionSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (c AddConditionSubroutine) GetName() string {
	return "addCondition"
}

func (c AddConditionSubroutine) Finalizers() []string {
	return []string{}
}

type FailureScenarioSubroutine struct {
	Retry              bool
	RequeAfter         bool
	FinalizeRetry      bool
	FinalizeRequeAfter bool
}

func (f FailureScenarioSubroutine) Process(_ context.Context, _ runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if f.RequeAfter {
		return controllerruntime.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return controllerruntime.Result{}, errors.NewOperatorError(fmt.Errorf("FailureScenarioSubroutine"), f.Retry, false)
}

func (f FailureScenarioSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if f.Retry {
		return controllerruntime.Result{Requeue: true}, nil
	}

	if f.RequeAfter {
		return controllerruntime.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return controllerruntime.Result{}, errors.NewOperatorError(fmt.Errorf("FailureScenarioSubroutine"), true, false)
}

func (f FailureScenarioSubroutine) Finalizers() []string {
	return []string{FailureScenarioSubroutineFinalizer}
}

func (c FailureScenarioSubroutine) GetName() string {
	return "FailureScenarioSubroutine"
}

type ImplementConditionsAndSpreadReconciles struct {
	testSupport.TestApiObject `json:",inline"`
}

func (m *ImplementConditionsAndSpreadReconciles) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}
func (m *ImplementConditionsAndSpreadReconciles) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}
func (m *ImplementConditionsAndSpreadReconciles) GetGeneration() int64 {
	return m.Generation
}
func (m *ImplementConditionsAndSpreadReconciles) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}
func (m *ImplementConditionsAndSpreadReconciles) SetObservedGeneration(g int64) {
	m.Status.ObservedGeneration = g
}

func (m *ImplementConditionsAndSpreadReconciles) GetNextReconcileTime() metav1.Time {
	return m.Status.NextReconcileTime
}
func (m *ImplementConditionsAndSpreadReconciles) SetNextReconcileTime(time metav1.Time) {
	m.Status.NextReconcileTime = time
}

type ContextValueSubroutine struct {
}

const ContextValueKey = keys.ContextKey("ContextValueKey")

func (f ContextValueSubroutine) Process(ctx context.Context, r runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if instance, ok := r.(*testSupport.TestApiObject); ok {
		instance.Status.Some = ctx.Value(ContextValueKey).(string)
	}
	return controllerruntime.Result{}, nil
}

func (f ContextValueSubroutine) Finalize(_ context.Context, _ runtimeobject.RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (f ContextValueSubroutine) Finalizers() []string {
	return []string{}
}

func (c ContextValueSubroutine) GetName() string {
	return "ContextValueSubroutine"
}
