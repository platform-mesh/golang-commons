// Package testresponse provides utilities for testing GraphQL API responses.
//
// It follows the same pattern as logger/testlogger - parse the response into a struct,
// then use standard testify assertions for flexible error checking.
//
// Basic usage:
//
//	gqlResp, err := testresponse.ParseGQLResponse(response)
//	require.NoError(t, err)
//
//	assert.True(t, gqlResp.HasErrors())
//	assert.Equal(t, 2, gqlResp.ErrorCount())
//	assert.True(t, gqlResp.HasErrorContaining("validation failed"))
//
// For simple success assertions in apitest chains:
//
//	suite.GqlApiTest(client).
//	    GraphQLQuery(query, vars).
//	    Expect(suite.T()).
//	    Status(http.StatusOK).
//	    Assert(testresponse.NoGQLErrors()).
//	    End()
package testresponse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GQLError represents a single GraphQL error from the response.
// It contains the error message and optional path and extensions metadata.
type GQLError struct {
	// Message is the human-readable error description.
	Message string `json:"message"`

	// Path indicates the response field that caused the error.
	// For example: ["mutation", "createUser", "input", "email"]
	Path []interface{} `json:"path,omitempty"`

	// Extensions contains additional error metadata like error codes.
	// For example: {"code": "VALIDATION_ERROR", "field": "email"}
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GQLResponse represents a parsed GraphQL response for testing.
// Use ParseGQLResponse to create instances, then assert with standard testify assertions.
type GQLResponse struct {
	// Errors contains all GraphQL errors from the response.
	// Access individual errors via index: gqlResp.Errors[0].Message
	Errors []GQLError `json:"errors,omitempty"`
}

// ParseGQLResponse extracts GraphQL errors from an HTTP response body.
// The response body is restored after reading so it can be read again by other assertions.
//
// Example:
//
//	gqlResp, err := testresponse.ParseGQLResponse(response)
//	require.NoError(t, err)
//	assert.True(t, gqlResp.HasErrorContaining("not found"))
func ParseGQLResponse(response *http.Response) (*GQLResponse, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	// Restore the body for subsequent reads
	response.Body = io.NopCloser(bytes.NewBuffer(body))

	var gqlResp GQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return &gqlResp, nil
}

// HasErrors returns true if the response contains any errors.
//
// Example:
//
//	if gqlResp.HasErrors() {
//	    t.Log("Response contained errors:", gqlResp.ErrorMessages())
//	}
func (r *GQLResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorCount returns the number of errors in the response.
//
// Example:
//
//	assert.Equal(t, 2, gqlResp.ErrorCount())
func (r *GQLResponse) ErrorCount() int {
	return len(r.Errors)
}

// ErrorMessages returns all error messages as a slice.
// Useful with testify's Contains or ElementsMatch assertions.
//
// Example:
//
//	assert.Contains(t, gqlResp.ErrorMessages(), "name is required")
//	assert.ElementsMatch(t, expectedErrors, gqlResp.ErrorMessages())
func (r *GQLResponse) ErrorMessages() []string {
	messages := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		messages[i] = err.Message
	}
	return messages
}

// HasErrorMessage returns true if any error has the exact message.
// For partial matching, use HasErrorContaining instead.
//
// Example:
//
//	assert.True(t, gqlResp.HasErrorMessage("user not found"))
func (r *GQLResponse) HasErrorMessage(expected string) bool {
	for _, err := range r.Errors {
		if err.Message == expected {
			return true
		}
	}
	return false
}

// HasErrorContaining returns true if any error message contains the substring.
// This checks ALL errors, not just the first one.
//
// Example:
//
//	assert.True(t, gqlResp.HasErrorContaining("validation failed"))
func (r *GQLResponse) HasErrorContaining(substring string) bool {
	for _, err := range r.Errors {
		if strings.Contains(err.Message, substring) {
			return true
		}
	}
	return false
}

// NoGQLErrors returns an assertion function for use in apitest chains.
// It returns an error if the response contains any GraphQL errors.
//
// Example:
//
//	suite.GqlApiTest(client).
//	    GraphQLQuery(query, vars).
//	    Expect(suite.T()).
//	    Status(http.StatusOK).
//	    Assert(testresponse.NoGQLErrors()).
//	    End()
func NoGQLErrors() func(response *http.Response, request *http.Request) error {
	return func(response *http.Response, request *http.Request) error {
		gqlResp, err := ParseGQLResponse(response)
		if err != nil {
			return err
		}
		if gqlResp.HasErrors() {
			return fmt.Errorf("expected no GraphQL errors but found %d: %v", gqlResp.ErrorCount(), gqlResp.ErrorMessages())
		}
		return nil
	}
}
