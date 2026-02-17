package testresponse_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/platform-mesh/golang-commons/graphql/testresponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This example demonstrates parsing a GraphQL response and checking for errors.
func ExampleParseGQLResponse() {
	t := &testing.T{} // In real tests, this comes from the test function

	body := `{"errors": [{"message": "validation failed"}, {"message": "not found"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, err := testresponse.ParseGQLResponse(response)
	require.NoError(t, err)

	assert.True(t, gqlResp.HasErrors())
	assert.Equal(t, 2, gqlResp.ErrorCount())
	assert.Contains(t, gqlResp.ErrorMessages(), "validation failed")
}

// HasErrorContaining checks ALL errors, not just the first one.
func ExampleGQLResponse_HasErrorContaining() {
	t := &testing.T{}

	body := `{"errors": [{"message": "first error"}, {"message": "second: not found"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Finds "not found" in the SECOND error - checks all errors, not just first
	assert.True(t, gqlResp.HasErrorContaining("not found"))
}

// HasErrorMessage requires an exact match, unlike HasErrorContaining.
func ExampleGQLResponse_HasErrorMessage() {
	t := &testing.T{}

	body := `{"errors": [{"message": "user not found"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Exact match works
	assert.True(t, gqlResp.HasErrorMessage("user not found"))

	// Partial does NOT match - use HasErrorContaining for that
	assert.False(t, gqlResp.HasErrorMessage("not found"))
}
