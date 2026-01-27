package testSupport

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type TestApiObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TestStatus `json:"status,omitempty"`
}
type TestStatus struct {
	Some               string             `json:"some,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	NextReconcileTime  metav1.Time        `json:"nextReconcileTime,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

func (t *TestApiObject) DeepCopyObject() runtime.Object {
	if c := t.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (t *TestApiObject) DeepCopy() *TestApiObject {
	if t == nil {
		return nil
	}
	out := new(TestApiObject)
	t.DeepCopyInto(out)
	return out
}
func (m *TestApiObject) DeepCopyInto(out *TestApiObject) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = m.Status
	if m.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(m.Status.Conditions))
		for i := range m.Status.Conditions {
			m.Status.Conditions[i].DeepCopyInto(&out.Status.Conditions[i])
		}
	}
}

type TestNoStatusApiObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (t *TestNoStatusApiObject) DeepCopyObject() runtime.Object {
	if c := t.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (t *TestNoStatusApiObject) DeepCopy() *TestNoStatusApiObject {
	if t == nil {
		return nil
	}
	out := new(TestNoStatusApiObject)
	t.DeepCopyInto(out)
	return out
}
func (m *TestNoStatusApiObject) DeepCopyInto(out *TestNoStatusApiObject) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
}
