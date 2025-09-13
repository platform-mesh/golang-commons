package local_middleware

import (
	"net/http"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/platform-mesh/golang-commons/context"
)

func LocalMiddleware(tenantId string, userId string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			ctx := request.Context()

			claims := &jwt.RegisteredClaims{Issuer: "localhost:8080", Subject: userId, Audience: jwt.ClaimStrings{"testing"}}
			token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
			if err != nil {
				panic(err) // This shouldn't happen, and if it does, only locally
			}

			ctx = context.AddWebTokenToContext(ctx, token, []jose.SignatureAlgorithm{"none"})
			ctx = context.AddTenantToContext(ctx, tenantId)

			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}
