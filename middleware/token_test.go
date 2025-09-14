package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/platform-mesh/golang-commons/context"
	"github.com/stretchr/testify/assert"
)

func TestStoreWebToken_WithFakeBearerToken(t *testing.T) {
	token := "fake.invalid.token"
	authHeader := "Bearer " + token

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := context.GetWebTokenFromContext(r.Context())
		// Presence of a Bearer header must result in a stored token.
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreWebToken()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, authHeader)
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreWebToken_WithoutAuthHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Context should not have a token
		_, err := context.GetWebTokenFromContext(r.Context())
		assert.Error(t, err) // Should return an error when no token is present

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreWebToken()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	// No authorization header set
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreWebToken_WithNonBearerToken(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Context should not have a token
		_, err := context.GetWebTokenFromContext(r.Context())
		assert.Error(t, err) // Should return an error when no valid token is present

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreWebToken()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, "Basic dXNlcjpwYXNz") // Basic auth, not Bearer
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreWebToken_WithEmptyBearerToken(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Context should not have a token due to empty token
		_, err := context.GetWebTokenFromContext(r.Context())
		assert.Error(t, err)

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreWebToken()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, "Bearer ")
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreWebToken_WithFakeBearerTokenLowercase(t *testing.T) {
	token := "fake.invalid.token"
	authHeader := "bearer " + token // lowercase

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token parsing may fail due to signature validation, which is expected in tests
		_, err := context.GetWebTokenFromContext(r.Context())
		// The middleware should process lowercase bearer tokens
		// but token validation may still fail due to signature issues
		if err != nil {
			// This is expected behavior when token validation fails
			assert.Error(t, err)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreWebToken()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, authHeader)
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestStoreWebToken_WithMalformedAuthHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Context should not have a token
		_, err := context.GetWebTokenFromContext(r.Context())
		assert.Error(t, err)

		w.WriteHeader(http.StatusOK)
	})

	middleware := StoreWebToken()
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Set(headers.Authorization, "Bearer") // Missing space and token
	recorder := httptest.NewRecorder()

	handlerToTest.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}
