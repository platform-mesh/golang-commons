package policy_services

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	pmcontext "github.com/platform-mesh/golang-commons/context"
)

func TestLocalTenantReader(t *testing.T) {
	t.Run("gets a tenant id", func(t *testing.T) {
		testContext := context.Background()

		// Arrange
		tenantId := "01emp2m3v3batersxj73qhm5zq"
		reader := NewLocalTenantRetriever(tenantId)

		claims := &jwt.RegisteredClaims{Issuer: "an issuer", Audience: jwt.ClaimStrings{"an audience"}}
		token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
		assert.NoError(t, err)

		testContext = pmcontext.AddWebTokenToContext(testContext, token, []jose.SignatureAlgorithm{jose.SignatureAlgorithm("none")})
		testContext = pmcontext.AddAuthHeaderToContext(testContext, fmt.Sprintf("Bearer %s", token))

		// Act
		id, err := reader.RetrieveTenant(testContext)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, tenantId, id)
	})
}
