package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/platform-mesh/golang-commons/context"
	"github.com/stretchr/testify/assert"
)

func TestStoreAuthHeader_WithAuthHeader(t *testing.T) {
	expectedAuth := "Bearer token123"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header is stored in context
		authFromContext, err := context.GetAuthHeaderFromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, expectedAuth, authFromContext)

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreAuthHeader()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, expectedAuth)
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreAuthHeader_WithoutAuthHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify empty auth header returns error when not set
		_, err := context.GetAuthHeaderFromContext(r.Context())
		assert.Error(t, err) // Should return error when no auth header is set

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreAuthHeader()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	// No authorization header set
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreAuthHeader_WithEmptyAuthHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify empty auth header returns error when empty
		_, err := context.GetAuthHeaderFromContext(r.Context())
		assert.Error(t, err) // Should return error when auth header is empty

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreAuthHeader()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, "")
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreAuthHeader_MultipleAuthHeaders(t *testing.T) {
	// Test behavior when multiple authorization headers are present
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should get the first/combined value
		authFromContext, err := context.GetAuthHeaderFromContext(r.Context())
		assert.NoError(t, err)
		assert.NotEmpty(t, authFromContext)

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreAuthHeader()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Add(headers.Authorization, "Bearer token1")
	req.Header.Add(headers.Authorization, "Bearer token2")
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
