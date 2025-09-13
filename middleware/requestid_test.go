package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetRequestIdWithIncomingHeader(t *testing.T) {

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := GetRequestId(r.Context())
		assert.Equal(t, "123", val)
	})

	// create the handler to test, using our custom "next" handler
	handlerToTest := SetRequestId()(nextHandler)

	// create a mock request to use
	req := httptest.NewRequest("GET", "http://testing", nil)
	req.Header.Add("X-Request-Id", "123")

	// call the handler using a mock response recorder (we'll not use that anyway)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}

func TestSetRequestIdWitoutIncomingHeader(t *testing.T) {

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := GetRequestId(r.Context())
		assert.Len(t, val, 36)
	})

	// create the handler to test, using our custom "next" handler
	handlerToTest := SetRequestId()(nextHandler)

	// create a mock request to use
	req := httptest.NewRequest("GET", "http://testing", nil)

	// call the handler using a mock response recorder (we'll not use that anyway)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}
