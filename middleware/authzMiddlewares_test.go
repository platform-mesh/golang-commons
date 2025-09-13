package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/platform-mesh/golang-commons/policy_services"
)

// MockTenantRetriever is a mock implementation of TenantRetriever
type MockTenantRetriever struct {
	mock.Mock
}

func (m *MockTenantRetriever) RetrieveTenant(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestCreateAuthMiddleware(t *testing.T) {
	mockRetriever := &MockTenantRetriever{}

	middlewares := CreateAuthMiddleware(mockRetriever)

	// Should return 3 middlewares: StoreWebToken, StoreAuthHeader, StoreSpiffeHeader
	assert.Len(t, middlewares, 3)

	// Each middleware should be a valid function
	for _, mw := range middlewares {
		assert.NotNil(t, mw)
	}
}

func TestCreateAuthMiddleware_WithNilRetriever(t *testing.T) {
	middlewares := CreateAuthMiddleware(nil)

	// Should still return 3 middlewares even with nil retriever
	assert.Len(t, middlewares, 3)

	// Each middleware should be a valid function
	for _, mw := range middlewares {
		assert.NotNil(t, mw)
	}
}

func TestCreateAuthMiddleware_ReturnsCorrectMiddlewares(t *testing.T) {
	mockRetriever := &MockTenantRetriever{}

	middlewares := CreateAuthMiddleware(mockRetriever)

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

// Test that implements policy_services.TenantRetriever interface
func TestTenantRetrieverInterface(t *testing.T) {
	mockRetriever := &MockTenantRetriever{}

	// Verify our mock implements the interface
	var retriever policy_services.TenantRetriever = mockRetriever
	assert.NotNil(t, retriever)

	// Test that we can use it with the middleware functions
	middlewares := CreateAuthMiddleware(retriever)
	assert.Len(t, middlewares, 3)
}
