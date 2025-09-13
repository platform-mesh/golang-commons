package middleware

import (
	"net/http"
)

// CreateAuthMiddleware returns a slice of HTTP middleware constructors that populate
// authentication-related request context and headers.
//
// The returned middlewares, applied in order, are:
// 1. StoreWebToken()
// 2. StoreAuthHeader()
// 3. StoreSpiffeHeader()
func CreateAuthMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		StoreWebToken(),
		StoreAuthHeader(),
		StoreSpiffeHeader(),
	}
}
