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
type GQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GQLResponse represents a parsed GraphQL response for testing.
// Use ParseGQLResponse to create instances, then assert with standard testify assertions.
type GQLResponse struct {
	Errors []GQLError `json:"errors,omitempty"`
}

// ParseGQLResponse extracts GraphQL errors from an HTTP response body.
// The response body is restored after reading so it can be read again by other assertions.
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
func (r *GQLResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorCount returns the number of errors in the response.
func (r *GQLResponse) ErrorCount() int {
	return len(r.Errors)
}

// ErrorMessages returns all error messages as a slice.
func (r *GQLResponse) ErrorMessages() []string {
	messages := make([]string, len(r.Errors))
	for i, err := range r.Errors {
		messages[i] = err.Message
	}
	return messages
}

// HasErrorMessage returns true if any error has the exact message.
func (r *GQLResponse) HasErrorMessage(expected string) bool {
	for _, err := range r.Errors {
		if err.Message == expected {
			return true
		}
	}
	return false
}

// HasErrorContaining returns true if any error message contains the substring.
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
