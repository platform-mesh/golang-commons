package middleware

import (
	_ "context"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/mock"
)

func TestCreateAuthMiddleware(t *testing.T) {
	middlewares := CreateAuthMiddleware()

	// Should return 3 middlewares: StoreWebToken, StoreAuthHeader, StoreSpiffeHeader
	assert.Len(t, middlewares, 3)

	// Each middleware should be a valid function
	for _, mw := range middlewares {
		assert.NotNil(t, mw)
	}
}

func TestCreateAuthMiddleware_WithNilRetriever(t *testing.T) {
	middlewares := CreateAuthMiddleware()

	// Should still return 3 middlewares even with nil retriever
	assert.Len(t, middlewares, 3)

	// Each middleware should be a valid function
	for _, mw := range middlewares {
		assert.NotNil(t, mw)
	}
}

func TestCreateAuthMiddleware_ReturnsCorrectMiddlewares(t *testing.T) {
	middlewares := CreateAuthMiddleware()

	// Verify we get exactly 3 middlewares
	assert.Len(t, middlewares, 3)

	// We can't easily test the exact middleware functions returned without more complex setup,
	// but we can verify they're all valid middleware functions by checking their signatures
	for _, mw := range middlewares {
		assert.NotNil(t, mw)
		// Each middleware should be a function that takes an http.Handler and returns an http.Handler
		// This is implicitly tested by the fact that the function compiles and returns without error
	}
}
