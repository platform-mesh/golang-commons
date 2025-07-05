package testSupport

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type FakeManager struct{ Client client.Client }

func (f *FakeManager) GetCluster(ctx context.Context, clusterName string) (cluster.Cluster, error) {
	return &FakeCluster{client: f.Client}, nil
}

var _ cluster.Cluster = (*FakeCluster)(nil)

type FakeCluster struct{ client client.Client }

func (f FakeCluster) GetHTTPClient() *http.Client {
	return nil
}

func (f FakeCluster) GetConfig() *rest.Config {
	return nil
}

func (f FakeCluster) GetCache() cache.Cache {
	return nil
}

func (f FakeCluster) GetScheme() *runtime.Scheme {
	return nil
}

func (f FakeCluster) GetClient() client.Client {
	return f.client
}

func (f FakeCluster) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (f FakeCluster) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

func (f FakeCluster) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (f FakeCluster) GetAPIReader() client.Reader {
	return nil
}

func (f FakeCluster) Start(ctx context.Context) error {
	return nil
}
