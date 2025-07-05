package testSupport

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type DummyRuntimeObject struct{}

func (d DummyRuntimeObject) GetObjectKind() schema.ObjectKind              { return nil }
func (d DummyRuntimeObject) DeepCopyObject() runtime.Object                { return nil }
func (d DummyRuntimeObject) GetAnnotations() map[string]string             { return nil }
func (d DummyRuntimeObject) SetAnnotations(annotations map[string]string)  {}
func (d DummyRuntimeObject) GetCreationTimestamp() metav1.Time             { return metav1.Time{} }
func (d DummyRuntimeObject) SetCreationTimestamp(timestamp metav1.Time)    {}
func (d DummyRuntimeObject) GetNamespace() string                          { return "" }
func (d DummyRuntimeObject) SetNamespace(namespace string)                 {}
func (d DummyRuntimeObject) GetName() string                               { return "" }
func (d DummyRuntimeObject) SetName(name string)                           {}
func (d DummyRuntimeObject) GetGenerateName() string                       { return "" }
func (d DummyRuntimeObject) SetGenerateName(name string)                   {}
func (d DummyRuntimeObject) GetUID() types.UID                             { return "" }
func (d DummyRuntimeObject) SetUID(uid types.UID)                          {}
func (d DummyRuntimeObject) GetGeneration() int64                          { return 0 }
func (d DummyRuntimeObject) SetGeneration(generation int64)                {}
func (d DummyRuntimeObject) GetResourceVersion() string                    { return "" }
func (d DummyRuntimeObject) SetResourceVersion(version string)             {}
func (d DummyRuntimeObject) GetFinalizers() []string                       { return nil }
func (d DummyRuntimeObject) SetFinalizers(finalizers []string)             {}
func (d DummyRuntimeObject) GetLabels() map[string]string                  { return nil }
func (d DummyRuntimeObject) SetLabels(labels map[string]string)            {}
func (d DummyRuntimeObject) GetOwnerReferences() []metav1.OwnerReference   { return nil }
func (d DummyRuntimeObject) SetOwnerReferences([]metav1.OwnerReference)    {}
func (d DummyRuntimeObject) GetManagedFields() []metav1.ManagedFieldsEntry { return nil }
func (d DummyRuntimeObject) SetManagedFields([]metav1.ManagedFieldsEntry)  {}
func (d DummyRuntimeObject) GetSelfLink() string                           { return "" }
func (d DummyRuntimeObject) SetSelfLink(selfLink string)                   {}
func (d DummyRuntimeObject) GetClusterName() string                        { return "" }
func (d DummyRuntimeObject) SetClusterName(clusterName string)             {}
func (d DummyRuntimeObject) GetDeletionTimestamp() *metav1.Time            { return nil }
func (d DummyRuntimeObject) SetDeletionTimestamp(*metav1.Time)             {}
func (d DummyRuntimeObject) GetDeletionGracePeriodSeconds() *int64         { return nil }
func (d DummyRuntimeObject) SetDeletionGracePeriodSeconds(*int64)          {}

type DummyRuntimeObjectWithConditions struct {
	DummyRuntimeObject
}

func (d DummyRuntimeObjectWithConditions) GetConditions() []metav1.Condition { return nil }
func (d DummyRuntimeObjectWithConditions) SetConditions([]metav1.Condition)  {}
