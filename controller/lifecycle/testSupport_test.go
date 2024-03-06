package lifecycle

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmfp/golang-commons/errors"
)

type testApiObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TestStatus `json:"status,omitempty"`
}

func (t *testApiObject) DeepCopyObject() runtime.Object {
	if c := t.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (t *testApiObject) DeepCopy() *testApiObject {
	if t == nil {
		return nil
	}
	out := new(testApiObject)
	t.DeepCopyInto(out)
	return out
}
func (t *testApiObject) DeepCopyInto(out *testApiObject) {
	*out = *t
	out.TypeMeta = t.TypeMeta
	t.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = t.Status
}

type changeStatusSubroutine struct {
	client client.Client
}

func (c changeStatusSubroutine) Process(_ context.Context, runtimeObj RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	instance := runtimeObj.(*testApiObject)
	instance.Status.Some = "other string"
	return controllerruntime.Result{}, nil
}

func (c changeStatusSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	//TODO implement me
	panic("implement me")
}

func (c changeStatusSubroutine) GetName() string {
	return "changeStatus"
}

func (c changeStatusSubroutine) Finalizers() []string {
	return []string{}
}

type TestStatus struct {
	Some string
}
