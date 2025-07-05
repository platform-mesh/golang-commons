package testSupport

import (
	"github.com/platform-mesh/golang-commons/controller/lifecycle/api"
	"github.com/platform-mesh/golang-commons/controller/lifecycle/subroutine"
	"github.com/platform-mesh/golang-commons/logger"
)

type TestLifecycleManager struct {
	Logger *logger.Logger
}

func (t TestLifecycleManager) Config() api.Config {
	return api.Config{
		ControllerName: "test-controller",
		OperatorName:   "test-operator",
		ReadOnly:       false,
	}
}
func (t TestLifecycleManager) Log() *logger.Logger                        { return t.Logger }
func (t TestLifecycleManager) Spreader() api.SpreadManager                { return nil }
func (t TestLifecycleManager) ConditionsManager() api.ConditionManager    { return nil }
func (t TestLifecycleManager) PrepareContextFunc() api.PrepareContextFunc { return nil }
func (t TestLifecycleManager) Subroutines() []subroutine.Subroutine       { return []subroutine.Subroutine{} }
