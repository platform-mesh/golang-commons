package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/platform-mesh/golang-commons/logger/testlogger"
	"github.com/stretchr/testify/assert"
)

func TestCreateMiddleware(t *testing.T) {
	log := testlogger.New()
	middlewares := CreateMiddleware(log.Logger)

	// Should return 5 middlewares
	assert.Len(t, middlewares, 5)

	// Each middleware should be a valid function
	for _, mw := range middlewares {
		assert.NotNil(t, mw)
	}

	// Test that middlewares can be chained
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply all middlewares
	var finalHandler http.Handler = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}

	req := httptest.NewRequest("GET", "http://testing", nil)
	recorder := httptest.NewRecorder()

	finalHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
