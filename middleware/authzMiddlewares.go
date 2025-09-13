package middleware

import (
	"net/http"

	"github.com/platform-mesh/golang-commons/policy_services"
)

// Deprecated: CreateMiddlewares use CreateAuthMiddleware instead.
func CreateMiddlewares(retriever policy_services.TenantRetriever) []func(http.Handler) http.Handler {
	return CreateAuthMiddleware(retriever)
}

func CreateAuthMiddleware(retriever policy_services.TenantRetriever) []func(http.Handler) http.Handler {
	mws := make([]func(http.Handler) http.Handler, 0, 5)

	mws = append(mws, StoreWebToken())
	mws = append(mws, StoreAuthHeader())
	mws = append(mws, StoreSpiffeHeader())

	return mws
}
