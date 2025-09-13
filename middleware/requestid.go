package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/platform-mesh/golang-commons/context/keys"
	"github.com/platform-mesh/golang-commons/logger"
)

const requestIdHeader = "X-Request-Id"

func SetRequestId() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			var requestId string
			if ids, ok := request.Header[requestIdHeader]; ok && len(ids) == 1 {
				requestId = ids[0]
			} else {
				// Generate a new request id, header was not received.
				requestId = uuid.New().String()
			}
			ctx = context.WithValue(ctx, keys.RequestIdCtxKey, requestId)
			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}

func SetRequestIdInLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			log := logger.LoadLoggerFromContext(ctx)
			log = logger.NewRequestLoggerFromZerolog(ctx, log.Logger)
			ctx = logger.SetLoggerInContext(ctx, log)
			next.ServeHTTP(responseWriter, request.WithContext(ctx))
		})
	}
}

func GetRequestId(ctx context.Context) string {
	if val, ok := ctx.Value(keys.RequestIdCtxKey).(string); ok {
		return val
	}
	return ""
}
