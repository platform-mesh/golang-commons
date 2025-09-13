package middleware

import (
	"net/http"

	"github.com/platform-mesh/golang-commons/logger"
)

// attaches a request-scoped logger (using the provided logger), assigns a request ID, and propagates that ID into the logger.
func CreateMiddleware(log *logger.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		SetOtelTracingContext(),
		SentryRecoverer,
		StoreLoggerMiddleware(log),
		SetRequestId(),
		SetRequestIdInLogger(),
	}
}
