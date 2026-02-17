# GraphQL Test Response

Package `testresponse` provides utilities for testing GraphQL API responses. It follows the same pattern as `logger/testlogger` - parse the response into a struct, then use standard testify assertions.

## Installation

```go
import "github.com/platform-mesh/golang-commons/graphql/testresponse"
```

## Usage

### Parsing Responses

```go
func (suite *TestSuite) TestMutation_ValidationErrors() {
    var response *http.Response
    suite.GqlApiTest(client).
        GraphQLQuery(mutation, vars).
        Expect(suite.T()).
        Status(http.StatusOK).
        Assert(func(resp *http.Response, _ *http.Request) error {
            response = resp
            return nil
        }).
        End()

    // Parse and assert
    gqlResp, err := testresponse.ParseGQLResponse(response)
    suite.Require().NoError(err)

    suite.True(gqlResp.HasErrors())
    suite.Equal(2, gqlResp.ErrorCount())
    suite.True(gqlResp.HasErrorContaining("validation"))
    suite.Contains(gqlResp.ErrorMessages(), "name is required")
}
```

### Simple Success Assertion

For tests that just need to verify no errors occurred:

```go
suite.GqlApiTest(client).
    GraphQLQuery(query, vars).
    Expect(suite.T()).
    Status(http.StatusOK).
    Assert(testresponse.NoGQLErrors()).
    End()
```

## API Reference

### ParseGQLResponse

```go
func ParseGQLResponse(response *http.Response) (*GQLResponse, error)
```

Parses an HTTP response body into a `GQLResponse` struct. The response body is restored after reading, allowing subsequent assertions to read it again.

### GQLResponse Methods

| Method | Return | Description |
|--------|--------|-------------|
| `HasErrors()` | `bool` | Returns true if response contains any errors |
| `ErrorCount()` | `int` | Returns the number of errors |
| `ErrorMessages()` | `[]string` | Returns all error messages as a slice |
| `HasErrorMessage(msg)` | `bool` | Returns true if any error matches exactly |
| `HasErrorContaining(sub)` | `bool` | Returns true if any error contains substring |

### GQLError Fields

```go
type GQLError struct {
    Message    string                 // Error message
    Path       []interface{}          // GraphQL path to the error
    Extensions map[string]interface{} // Additional error metadata
}
```

### NoGQLErrors

```go
func NoGQLErrors() func(*http.Response, *http.Request) error
```

Returns an assertion function for use in apitest chains. Fails if the response contains any GraphQL errors.

## Why This Pattern?

Instead of custom assertion functions that only check the first error:

```go
// Limited - only checks first error
Assert(HasGQLErrorContaining("not found"))
```

The struct-based approach provides flexibility:

```go
// Flexible - check any error, use familiar assertions
gqlResp, _ := testresponse.ParseGQLResponse(response)
assert.True(t, gqlResp.HasErrorContaining("not found"))
assert.Equal(t, 2, gqlResp.ErrorCount())
assert.Contains(t, gqlResp.ErrorMessages(), "specific error")
```

Benefits:
- Check ANY error, not just the first one
- Use standard testify assertions with better error messages
- Access Path and Extensions for detailed validation
- Consistent with `testlogger` pattern
