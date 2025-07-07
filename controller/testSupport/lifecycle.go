package testSupport

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/platform-mesh/golang-commons/controller/lifecycle/api"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/runtimeobject"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	"github.com/platform-mesh/golang-commons/logger"
)

type TestLifecycleManager struct {
	Logger            *logger.Logger
	SubroutinesArr    []subroutine.Subroutine
	spreader          api.SpreadManager
	conditionsManager api.ConditionManager
	ShouldReconcile   bool
}

func (t TestLifecycleManager) Config() api.Config {
	return api.Config{
		ControllerName: "test-controller",
		OperatorName:   "test-operator",
		ReadOnly:       false,
	}
}
func (t TestLifecycleManager) Log() *logger.Logger                        { return t.Logger }
func (t TestLifecycleManager) Spreader() api.SpreadManager                { return t.spreader }
func (t TestLifecycleManager) ConditionsManager() api.ConditionManager    { return t.conditionsManager }
func (t TestLifecycleManager) PrepareContextFunc() api.PrepareContextFunc { return nil }
func (t TestLifecycleManager) Subroutines() []subroutine.Subroutine       { return t.SubroutinesArr }
func (l *TestLifecycleManager) WithSpreadingReconciles() *TestLifecycleManager {
	l.spreader = &TestSpreader{ShouldReconcile: l.ShouldReconcile}
	return l
}
func (l *TestLifecycleManager) WithConditionManagement() *TestLifecycleManager {
	l.conditionsManager = &TestConditionManager{}
	return l
}

type TestSpreader struct {
	ShouldReconcile bool
}

func (t TestSpreader) ReconcileRequired(instance runtimeobject.RuntimeObject, log *logger.Logger) bool {
	return t.ShouldReconcile
}

func (t TestSpreader) ToRuntimeObjectSpreadReconcileStatusInterface(instance runtimeobject.RuntimeObject, log *logger.Logger) (api.RuntimeObjectSpreadReconcileStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestSpreader) MustToRuntimeObjectSpreadReconcileStatusInterface(instance runtimeobject.RuntimeObject, log *logger.Logger) api.RuntimeObjectSpreadReconcileStatus {

	//TODO implement me
	panic("implement me")
}

func (t TestSpreader) OnNextReconcile(instance runtimeobject.RuntimeObject, log *logger.Logger) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
}

func (t TestSpreader) RemoveRefreshLabelIfExists(instance runtimeobject.RuntimeObject) bool {
	return false
}

func (t TestSpreader) SetNextReconcileTime(instanceStatusObj api.RuntimeObjectSpreadReconcileStatus, log *logger.Logger) {
	instanceStatusObj.SetNextReconcileTime(metav1.NewTime(time.Now().Add(10 * time.Hour)))
}

func (t TestSpreader) UpdateObservedGeneration(instanceStatusObj api.RuntimeObjectSpreadReconcileStatus, log *logger.Logger) {
	instanceStatusObj.SetObservedGeneration(instanceStatusObj.GetGeneration())
}

type TestConditionManager struct{}

func (t TestConditionManager) MustToRuntimeObjectConditionsInterface(instance runtimeobject.RuntimeObject, log *logger.Logger) api.RuntimeObjectConditions {
	//TODO implement me
	panic("implement me")
}

func (t TestConditionManager) SetInstanceConditionUnknownIfNotSet(conditions *[]metav1.Condition) bool {
	//TODO implement me
	panic("implement me")
}

func (t TestConditionManager) SetSubroutineConditionToUnknownIfNotSet(conditions *[]metav1.Condition, subroutine subroutine.Subroutine, isFinalize bool, log *logger.Logger) bool {
	//TODO implement me
	panic("implement me")
}

func (t TestConditionManager) SetSubroutineCondition(conditions *[]metav1.Condition, subroutine subroutine.Subroutine, subroutineResult ctrl.Result, subroutineErr error, isFinalize bool, log *logger.Logger) bool {
	//TODO implement me
	panic("implement me")
}

func (t TestConditionManager) SetInstanceConditionReady(conditions *[]metav1.Condition, status metav1.ConditionStatus) bool {
	//TODO implement me
	panic("implement me")
}

func (t TestConditionManager) ToRuntimeObjectConditionsInterface(instance runtimeobject.RuntimeObject, log *logger.Logger) (api.RuntimeObjectConditions, error) {
	//TODO implement me
	panic("implement me")
}
