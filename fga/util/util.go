package util

import (
	"fmt"
	"strings"
)

const maxRelationLength = 50

func ConvertToTypeName(group, singular string) string {
	if group == "" {
		group = "core"
	}

	// Cap the length of the group_singular string
	objectType := capGroupSingularLength(group, singular, maxRelationLength)
	// Make sure the result does not start with an underscore
	objectType = strings.TrimPrefix(objectType, "_")
	// Replace dots with underscores in the final objectType
	objectType = strings.ReplaceAll(objectType, ".", "_")
	// Convert to lowercase
	objectType = strings.ToLower(objectType)
	return objectType
}

func capGroupSingularLength(group, singular string, maxLength int) string {
	groupKind := fmt.Sprintf("%s_%s", group, singular)
	maxRelation := fmt.Sprintf("create_%s", groupKind)

	if len(maxRelation) > maxLength && maxLength > 0 {
		truncateLen := len(maxRelation) - maxLength
		return groupKind[truncateLen:]
	}
	return groupKind
}
