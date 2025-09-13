package middleware

import (
	"net/http"
	"strings"

	"github.com/go-http-utils/headers"
	"github.com/go-jose/go-jose/v4"
	"github.com/platform-mesh/golang-commons/context"
)

const tokenAuthPrefix = "BEARER"

var SignatureAlgorithms = []jose.SignatureAlgorithm{jose.RS256}

// StoreWebToken returns middleware that extracts a JWT from the HTTP `Authorization` header
// and stores it in the request context for downstream handlers.
//
// The middleware looks for an Authorization header of the form `Bearer <token>` (scheme match is
// case-insensitive). When present, the token is added to the context via
// context.AddWebTokenToContext using the package's SignatureAlgorithms. If the header is absent,
// malformed, or not a Bearer token, the request context is left unchanged.
func StoreWebToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			auth := strings.Split(request.Header.Get(headers.Authorization), " ")
			if len(auth) > 1 && strings.ToUpper(auth[0]) == tokenAuthPrefix {
				ctx = context.AddWebTokenToContext(ctx, auth[1], SignatureAlgorithms)
			}

			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}
