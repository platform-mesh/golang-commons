package model

import (
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/platform-mesh/golang-commons/fga/util"
)

// BuildObjectType renders the canonical OpenFGA object type from an API group
// and singular resource name.
func BuildObjectType(group, singular string) string {
	return util.ConvertToTypeName(group, singular)
}

// BuildObjectName renders the canonical OpenFGA object name using the API group
// and singular resource name.
func BuildObjectName(group, singular, clusterID, name string, namespace *string) string {
	return BuildObjectNameFromType(BuildObjectType(group, singular), clusterID, name, namespace)
}

// BuildObjectNameFromType renders the canonical OpenFGA object name from a
// fully normalized OpenFGA object type.
func BuildObjectNameFromType(objectType, clusterID, name string, namespace *string) string {
	if namespace != nil && *namespace != "" {
		return fmt.Sprintf("%s:%s/%s/%s", objectType, clusterID, *namespace, name)
	}

	return fmt.Sprintf("%s:%s/%s", objectType, clusterID, name)
}

// BuildParentTuples renders the canonical parent hierarchy tuples used by the
// webhook, operator, and search consumers.
func BuildParentTuples(parentObject, object string, namespaceObject *string) []*openfgav1.TupleKey {
	if namespaceObject != nil && *namespaceObject != "" {
		return []*openfgav1.TupleKey{
			{
				Object:   *namespaceObject,
				Relation: "parent",
				User:     parentObject,
			},
			{
				Object:   object,
				Relation: "parent",
				User:     *namespaceObject,
			},
		}
	}

	if object == parentObject {
		return nil
	}

	return []*openfgav1.TupleKey{
		{
			Object:   object,
			Relation: "parent",
			User:     parentObject,
		},
	}
}
