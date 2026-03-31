package model

import "testing"

func TestBuildObjectType(t *testing.T) {
	tests := []struct {
		name     string
		group    string
		singular string
		want     string
	}{
		{
			name:     "custom resource",
			group:    "core.platform-mesh.io",
			singular: "account",
			want:     "core_platform-mesh_io_account",
		},
		{
			name:     "core resource",
			group:    "",
			singular: "namespace",
			want:     "core_namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildObjectType(tt.group, tt.singular); got != tt.want {
				t.Fatalf("BuildObjectType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildObjectName(t *testing.T) {
	namespace := "ns1"

	tests := []struct {
		name      string
		group     string
		singular  string
		clusterID string
		resource  string
		namespace *string
		want      string
	}{
		{
			name:      "namespaced resource",
			group:     "core.platform-mesh.io",
			singular:  "component",
			clusterID: "cluster1",
			resource:  "comp1",
			namespace: &namespace,
			want:      "core_platform-mesh_io_component:cluster1/ns1/comp1",
		},
		{
			name:      "cluster scoped resource",
			group:     "core.platform-mesh.io",
			singular:  "account",
			clusterID: "cluster1",
			resource:  "acc1",
			want:      "core_platform-mesh_io_account:cluster1/acc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildObjectName(tt.group, tt.singular, tt.clusterID, tt.resource, tt.namespace); got != tt.want {
				t.Fatalf("BuildObjectName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildParentTuples(t *testing.T) {
	namespaceObject := "core_namespace:cluster1/ns1"

	t.Run("namespaced", func(t *testing.T) {
		tuples := BuildParentTuples("core_platform-mesh_io_account:origin/account", "core_platform-mesh_io_component:cluster1/ns1/comp1", &namespaceObject)
		if len(tuples) != 2 {
			t.Fatalf("expected 2 tuples, got %d", len(tuples))
		}
		if tuples[0].Object != namespaceObject || tuples[0].User != "core_platform-mesh_io_account:origin/account" {
			t.Fatalf("unexpected first tuple: %+v", tuples[0])
		}
		if tuples[1].Object != "core_platform-mesh_io_component:cluster1/ns1/comp1" || tuples[1].User != namespaceObject {
			t.Fatalf("unexpected second tuple: %+v", tuples[1])
		}
	})

	t.Run("cluster scoped", func(t *testing.T) {
		tuples := BuildParentTuples("core_platform-mesh_io_account:origin/account", "core_platform-mesh_io_component:cluster1/comp1", nil)
		if len(tuples) != 1 {
			t.Fatalf("expected 1 tuple, got %d", len(tuples))
		}
		if tuples[0].Object != "core_platform-mesh_io_component:cluster1/comp1" || tuples[0].User != "core_platform-mesh_io_account:origin/account" {
			t.Fatalf("unexpected tuple: %+v", tuples[0])
		}
	})

	t.Run("self parent skipped", func(t *testing.T) {
		tuples := BuildParentTuples("core_platform-mesh_io_account:origin/account", "core_platform-mesh_io_account:origin/account", nil)
		if len(tuples) != 0 {
			t.Fatalf("expected 0 tuples, got %d", len(tuples))
		}
	})
}
