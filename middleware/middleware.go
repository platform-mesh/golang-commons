package middleware

import (
	"net/http"

	"github.com/platform-mesh/golang-commons/logger"
)

func CreateMiddleware(log *logger.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		SetOtelTracingContext(),
		SentryRecoverer,
		StoreLoggerMiddleware(log),
		SetRequestId(),
		SetRequestIdInLogger(),
	}
}
