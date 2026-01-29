package testSupport

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/platform-mesh/golang-commons/context/keys"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const FailureScenarioSubroutineFinalizer = "failuresubroutine"
const ChangeStatusSubroutineFinalizer = "changestatus"

type ImplementConditions struct {
	TestApiObject `json:",inline"`
}

func (m *ImplementConditions) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *ImplementConditions) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

type ImplementingSpreadReconciles struct {
	TestApiObject `json:",inline"`
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

	Status TestStatus `json:"status,omitempty"`
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

func (c ChangeStatusSubroutine) Process(
	_ context.Context,
	runtimeObj runtimeobject.RuntimeObject,
) subroutine.Result {
	if instance, ok := runtimeObj.(*TestApiObject); ok {
		instance.Status.Some = "other string"
	}
	if instance, ok := runtimeObj.(*ImplementingSpreadReconciles); ok {
		instance.Status.Some = "other string"
	}

	if instance, ok := runtimeObj.(*ImplementConditions); ok {
		instance.Status.Some = "other string"
	}
	return subroutine.OK()
}

func (c ChangeStatusSubroutine) Finalize(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	return subroutine.OK()
}

func (c ChangeStatusSubroutine) GetName() string {
	return "changeStatus"
}

func (c ChangeStatusSubroutine) Finalizers(instance runtimeobject.RuntimeObject) []string {
	return []string{"changestatus"}
}

type AddConditionSubroutine struct {
	Ready metav1.ConditionStatus
}

func (c AddConditionSubroutine) Process(
	_ context.Context,
	runtimeObj runtimeobject.RuntimeObject,
) subroutine.Result {
	if instance, ok := runtimeObj.(*ImplementConditions); ok {
		instance.Status.Some = "other string"
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:    "test",
			Status:  c.Ready,
			Reason:  "test",
			Message: "test",
		})
	}

	return subroutine.OK()
}

func (c AddConditionSubroutine) Finalize(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	return subroutine.OK()
}

func (c AddConditionSubroutine) GetName() string {
	return "addCondition"
}

func (c AddConditionSubroutine) Finalizers(instance runtimeobject.RuntimeObject) []string {
	return []string{}
}

type FailureScenarioSubroutine struct {
	Retry              bool
	RequeAfter         bool
	FinalizeRetry      bool
	FinalizeRequeAfter bool
}

func (f FailureScenarioSubroutine) Process(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	if f.RequeAfter {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: 10 * time.Second})
	}

	if f.Retry {
		return subroutine.Retry(fmt.Errorf("FailureScenarioSubroutine"))
	}
	return subroutine.Fail(fmt.Errorf("FailureScenarioSubroutine"))
}

func (f FailureScenarioSubroutine) Finalize(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	if f.RequeAfter {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: 10 * time.Second})
	}

	return subroutine.Retry(fmt.Errorf("FailureScenarioSubroutine"))
}

func (f FailureScenarioSubroutine) Finalizers(instance runtimeobject.RuntimeObject) []string {
	return []string{FailureScenarioSubroutineFinalizer}
}

func (c FailureScenarioSubroutine) GetName() string {
	return "FailureScenarioSubroutine"
}

type ImplementConditionsAndSpreadReconciles struct {
	TestApiObject `json:",inline"`
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

func (f ContextValueSubroutine) Process(
	ctx context.Context,
	r runtimeobject.RuntimeObject,
) subroutine.Result {
	if instance, ok := r.(*TestApiObject); ok {
		instance.Status.Some = ctx.Value(ContextValueKey).(string)
	}
	return subroutine.OK()
}

func (f ContextValueSubroutine) Finalize(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	return subroutine.OK()
}

func (f ContextValueSubroutine) Finalizers(instance runtimeobject.RuntimeObject) []string {
	return []string{}
}

func (c ContextValueSubroutine) GetName() string {
	return "ContextValueSubroutine"
}

type ChainStopSubroutine struct {
	Name         string
	StopChain    bool
	StopReason   string
	RequeueAfter time.Duration
	Processed    *bool
	Finalized    *bool
}

func (c ChainStopSubroutine) Process(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	if c.Processed != nil {
		*c.Processed = true
	}
	if c.StopChain {
		reason := c.StopReason
		if reason == "" {
			reason = "chain stop requested"
		}
		if c.RequeueAfter > 0 {
			return subroutine.StopWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter}, reason)
		}
		return subroutine.Stop(reason)
	}
	if c.RequeueAfter > 0 {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter})
	}
	return subroutine.OK()
}

func (c ChainStopSubroutine) Finalize(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	if c.Finalized != nil {
		*c.Finalized = true
	}
	if c.StopChain {
		reason := c.StopReason
		if reason == "" {
			reason = "chain stop requested"
		}
		if c.RequeueAfter > 0 {
			return subroutine.StopWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter}, reason)
		}
		return subroutine.Stop(reason)
	}
	if c.RequeueAfter > 0 {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter})
	}
	return subroutine.OK()
}

func (c ChainStopSubroutine) GetName() string {
	if c.Name != "" {
		return c.Name
	}
	return "ChainStopSubroutine"
}

func (c ChainStopSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	return []string{}
}

type ChainStopFinalizerSubroutine struct {
	Name          string
	FinalizerName string
	StopChain     bool
	StopReason    string
	RequeueAfter  time.Duration
	Processed     *bool
	Finalized     *bool
}

func (c ChainStopFinalizerSubroutine) Process(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	if c.Processed != nil {
		*c.Processed = true
	}
	if c.StopChain {
		reason := c.StopReason
		if reason == "" {
			reason = "chain stop requested"
		}
		if c.RequeueAfter > 0 {
			return subroutine.StopWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter}, reason)
		}
		return subroutine.Stop(reason)
	}
	if c.RequeueAfter > 0 {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter})
	}
	return subroutine.OK()
}

func (c ChainStopFinalizerSubroutine) Finalize(
	_ context.Context,
	_ runtimeobject.RuntimeObject,
) subroutine.Result {
	if c.Finalized != nil {
		*c.Finalized = true
	}
	if c.StopChain {
		reason := c.StopReason
		if reason == "" {
			reason = "chain stop requested"
		}
		if c.RequeueAfter > 0 {
			return subroutine.StopWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter}, reason)
		}
		return subroutine.Stop(reason)
	}
	if c.RequeueAfter > 0 {
		return subroutine.OKWithRequeue(controllerruntime.Result{RequeueAfter: c.RequeueAfter})
	}
	return subroutine.OK()
}

func (c ChainStopFinalizerSubroutine) GetName() string {
	if c.Name != "" {
		return c.Name
	}
	return "ChainStopFinalizerSubroutine"
}

func (c ChainStopFinalizerSubroutine) Finalizers(_ runtimeobject.RuntimeObject) []string {
	if c.FinalizerName != "" {
		return []string{c.FinalizerName}
	}
	return []string{"chainstopfinalizer"}
}
