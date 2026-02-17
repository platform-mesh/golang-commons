package testresponse_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/platform-mesh/golang-commons/graphql/testresponse"
)

func ExampleParseGQLResponse() {
	// Simulate a GraphQL response with errors
	body := `{
		"data": null,
		"errors": [
			{"message": "validation failed: name is required"},
			{"message": "validation failed: email is invalid"}
		]
	}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, err := testresponse.ParseGQLResponse(response)
	if err != nil {
		panic(err)
	}

	fmt.Println("Has errors:", gqlResp.HasErrors())
	fmt.Println("Error count:", gqlResp.ErrorCount())
	fmt.Println("First message:", gqlResp.Errors[0].Message)

	// Output:
	// Has errors: true
	// Error count: 2
	// First message: validation failed: name is required
}

func ExampleGQLResponse_HasErrorContaining() {
	body := `{"errors": [{"message": "user not found"}, {"message": "permission denied"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Check if any error contains a substring
	fmt.Println("Contains 'not found':", gqlResp.HasErrorContaining("not found"))
	fmt.Println("Contains 'denied':", gqlResp.HasErrorContaining("denied"))
	fmt.Println("Contains 'timeout':", gqlResp.HasErrorContaining("timeout"))

	// Output:
	// Contains 'not found': true
	// Contains 'denied': true
	// Contains 'timeout': false
}

func ExampleGQLResponse_HasErrorMessage() {
	body := `{"errors": [{"message": "exact error message"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Exact match required
	fmt.Println("Exact match:", gqlResp.HasErrorMessage("exact error message"))
	fmt.Println("Partial match:", gqlResp.HasErrorMessage("exact error"))

	// Output:
	// Exact match: true
	// Partial match: false
}

func ExampleGQLResponse_ErrorMessages() {
	body := `{"errors": [{"message": "error one"}, {"message": "error two"}]}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	// Get all messages as a slice - useful with slices.Contains
	messages := gqlResp.ErrorMessages()
	fmt.Println("Contains 'error one':", slices.Contains(messages, "error one"))
	fmt.Println("Contains 'error three':", slices.Contains(messages, "error three"))

	// Output:
	// Contains 'error one': true
	// Contains 'error three': false
}

func ExampleGQLResponse_accessingErrorDetails() {
	body := `{
		"errors": [{
			"message": "field validation failed",
			"path": ["mutation", "createUser", "input", "email"],
			"extensions": {"code": "VALIDATION_ERROR", "field": "email"}
		}]
	}`
	response := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}

	gqlResp, _ := testresponse.ParseGQLResponse(response)

	err := gqlResp.Errors[0]
	fmt.Println("Message:", err.Message)
	fmt.Println("Path length:", len(err.Path))
	fmt.Println("Extension code:", err.Extensions["code"])

	// Output:
	// Message: field validation failed
	// Path length: 4
	// Extension code: VALIDATION_ERROR
}

func ExampleNoGQLErrors() {
	// Success response - no errors
	successBody := `{"data": {"user": {"id": "123", "name": "John"}}}`
	successResp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(successBody)),
	}

	err := testresponse.NoGQLErrors()(successResp, nil)
	fmt.Println("Success response error:", err)

	// Error response
	errorBody := `{"errors": [{"message": "not found"}]}`
	errorResp := &http.Response{
		Body: io.NopCloser(bytes.NewBufferString(errorBody)),
	}

	err = testresponse.NoGQLErrors()(errorResp, nil)
	fmt.Println("Error response has error:", err != nil)

	// Output:
	// Success response error: <nil>
	// Error response has error: true
}
