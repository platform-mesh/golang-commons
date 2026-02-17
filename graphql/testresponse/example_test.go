package testresponse_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/platform-mesh/golang-commons/graphql/testresponse"
)

// This example demonstrates parsing a GraphQL response and checking for errors.
// In real tests, use testify assertions instead of fmt.Println.
func ExampleParseGQLResponse() {
	body := `{"errors": [{"message": "validation failed"}, {"message": "not found"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// In tests: assert.True(t, gqlResp.HasErrors())
	// In tests: assert.Equal(t, 2, gqlResp.ErrorCount())
	fmt.Println(gqlResp.HasErrors(), gqlResp.ErrorCount())

	// Output:
	// true 2
}

// HasErrorContaining checks ALL errors, not just the first one.
func ExampleGQLResponse_HasErrorContaining() {
	body := `{"errors": [{"message": "first error"}, {"message": "second: not found"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Finds "not found" in the SECOND error
	// In tests: assert.True(t, gqlResp.HasErrorContaining("not found"))
	fmt.Println(gqlResp.HasErrorContaining("not found"))

	// Output:
	// true
}

// HasErrorMessage requires an exact match, unlike HasErrorContaining.
func ExampleGQLResponse_HasErrorMessage() {
	body := `{"errors": [{"message": "user not found"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Exact match works
	// In tests: assert.True(t, gqlResp.HasErrorMessage("user not found"))
	fmt.Println(gqlResp.HasErrorMessage("user not found"))

	// Partial does NOT match
	// In tests: assert.False(t, gqlResp.HasErrorMessage("not found"))
	fmt.Println(gqlResp.HasErrorMessage("not found"))

	// Output:
	// true
	// false
}
