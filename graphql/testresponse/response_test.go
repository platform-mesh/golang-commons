package testresponse

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockResponse(body string) *http.Response {
	return &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestNoGQLErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantError bool
	}{
		{
			name:      "no errors field",
			body:      `{"data": {"user": {"id": "1"}}}`,
			wantError: false,
		},
		{
			name:      "empty errors array",
			body:      `{"data": {"user": {"id": "1"}}, "errors": []}`,
			wantError: false,
		},
		{
			name:      "with errors",
			body:      `{"data": null, "errors": [{"message": "not found"}]}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NoGQLErrors()(mockResponse(tt.body), nil)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseGQLResponse(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantErrorCount int
		wantMessages   []string
	}{
		{
			name:           "no errors",
			body:           `{"data": {"user": {"id": "1"}}}`,
			wantErrorCount: 0,
			wantMessages:   []string{},
		},
		{
			name:           "single error",
			body:           `{"errors": [{"message": "not found"}]}`,
			wantErrorCount: 1,
			wantMessages:   []string{"not found"},
		},
		{
			name:           "multiple errors",
			body:           `{"errors": [{"message": "error one"}, {"message": "error two"}, {"message": "error three"}]}`,
			wantErrorCount: 3,
			wantMessages:   []string{"error one", "error two", "error three"},
		},
		{
			name:           "error with path and extensions",
			body:           `{"errors": [{"message": "field error", "path": ["query", "user", "name"], "extensions": {"code": "VALIDATION_ERROR"}}]}`,
			wantErrorCount: 1,
			wantMessages:   []string{"field error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := mockResponse(tt.body)
			gqlResp, err := ParseGQLResponse(resp)
			require.NoError(t, err)

			assert.Equal(t, tt.wantErrorCount, gqlResp.ErrorCount())
			assert.Equal(t, tt.wantMessages, gqlResp.ErrorMessages())
			assert.Equal(t, tt.wantErrorCount > 0, gqlResp.HasErrors())
		})
	}
}

func TestGQLResponse_HasErrorMessage(t *testing.T) {
	body := `{"errors": [{"message": "first error"}, {"message": "second error"}]}`
	resp := mockResponse(body)
	gqlResp, err := ParseGQLResponse(resp)
	require.NoError(t, err)

	// Exact match - first error
	assert.True(t, gqlResp.HasErrorMessage("first error"))
	// Exact match - second error
	assert.True(t, gqlResp.HasErrorMessage("second error"))
	// No match
	assert.False(t, gqlResp.HasErrorMessage("third error"))
	// Partial doesn't match
	assert.False(t, gqlResp.HasErrorMessage("first"))
}

func TestGQLResponse_HasErrorContaining(t *testing.T) {
	body := `{"errors": [{"message": "user not found"}, {"message": "permission denied"}]}`
	resp := mockResponse(body)
	gqlResp, err := ParseGQLResponse(resp)
	require.NoError(t, err)

	// Substring in first error
	assert.True(t, gqlResp.HasErrorContaining("not found"))
	// Substring in second error
	assert.True(t, gqlResp.HasErrorContaining("denied"))
	// Substring not in any error
	assert.False(t, gqlResp.HasErrorContaining("timeout"))
	// Full message works too
	assert.True(t, gqlResp.HasErrorContaining("user not found"))
}

func TestGQLResponse_DirectAssertions(t *testing.T) {
	// Demonstrates the intended usage pattern with testify assertions
	body := `{"errors": [{"message": "validation failed: name required"}, {"message": "validation failed: email invalid"}]}`
	resp := mockResponse(body)
	gqlResp, err := ParseGQLResponse(resp)
	require.NoError(t, err)

	// Direct assertions on the struct
	assert.True(t, gqlResp.HasErrors())
	assert.Equal(t, 2, gqlResp.ErrorCount())
	assert.Contains(t, gqlResp.ErrorMessages(), "validation failed: name required")
	assert.Contains(t, gqlResp.ErrorMessages(), "validation failed: email invalid")
	assert.True(t, gqlResp.HasErrorContaining("validation failed"))

	// Can also access individual errors
	assert.Equal(t, "validation failed: name required", gqlResp.Errors[0].Message)
	assert.Equal(t, "validation failed: email invalid", gqlResp.Errors[1].Message)
}

func TestParseGQLResponse_RestoresBody(t *testing.T) {
	// Verify that the response body can be read again after parsing
	body := `{"errors": [{"message": "test error"}]}`
	resp := mockResponse(body)

	// Parse once
	gqlResp1, err := ParseGQLResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, 1, gqlResp1.ErrorCount())

	// Parse again - should work because body is restored
	gqlResp2, err := ParseGQLResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, 1, gqlResp2.ErrorCount())
}

func TestGQLError_PathAndExtensions(t *testing.T) {
	body := `{"errors": [{"message": "field error", "path": ["query", "user", 0, "name"], "extensions": {"code": "VALIDATION_ERROR", "field": "name"}}]}`
	resp := mockResponse(body)
	gqlResp, err := ParseGQLResponse(resp)
	require.NoError(t, err)

	require.Len(t, gqlResp.Errors, 1)
	gqlErr := gqlResp.Errors[0]

	assert.Equal(t, "field error", gqlErr.Message)
	assert.Len(t, gqlErr.Path, 4)
	assert.Equal(t, "VALIDATION_ERROR", gqlErr.Extensions["code"])
	assert.Equal(t, "name", gqlErr.Extensions["field"])
}
