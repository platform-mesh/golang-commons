package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/platform-mesh/golang-commons/logger"
)

// Recoverer implements a middleware that recover from panics, sends them to Sentry
// log the panic together with a stack trace and sends HTTP status 500
func SentryRecoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil && err != http.ErrAbortHandler {
				log := logger.LoadLoggerFromContext(r.Context())
				log.Error().Interface("panic", err).Interface("stack", debug.Stack()).Msg("recovered http panic")
				sentry.CurrentHub().Recover(err)
				sentry.Flush(time.Second * 5)

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
