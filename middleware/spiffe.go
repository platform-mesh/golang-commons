package middleware

import (
	"net/http"

	"github.com/platform-mesh/golang-commons/context"
	"github.com/platform-mesh/golang-commons/jwt"
)

// StoreSpiffeHeader returns an HTTP middleware that extracts a SPIFFE URL from the request headers
// and, if present, inserts it into the request context for downstream handlers.
//
// The middleware always calls the next handler; when a SPIFFE URL is found it updates the request's
// context with that value so subsequent handlers can retrieve it.
func StoreSpiffeHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			uriVal := jwt.GetSpiffeUrlValue(request.Header)

			if uriVal != nil {
				ctx = context.AddSpiffeToContext(ctx, *uriVal)
			}
			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}
