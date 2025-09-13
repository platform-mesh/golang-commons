package middleware

import (
	"net/http"
)

// Middleware defines a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// CreateAuthMiddleware returns a slice of Middleware functions for authentication and authorization.
// The returned middlewares are: StoreWebToken, StoreAuthHeader, and StoreSpiffeHeader.
func CreateAuthMiddleware() []Middleware {
	return []Middleware{
		StoreWebToken(),
		StoreAuthHeader(),
		StoreSpiffeHeader(),
	}
}
