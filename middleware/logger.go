package middleware

import (
	"net/http"

	"github.com/platform-mesh/golang-commons/logger"
)

// StoreLoggerMiddleware is a middleware that stores a given Logger in the request context
func StoreLoggerMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logger.SetLoggerInContext(r.Context(), log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
