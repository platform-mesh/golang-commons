package jwt

import (
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var signatureAlgorithms = []jose.SignatureAlgorithm{jose.HS256}
var joseTestKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes

func TestNew(t *testing.T) {
	issuer := "my-issuer"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer: issuer,
	})
	tokenString, err := token.SignedString(joseTestKey)
	assert.NoError(t, err)

	webToken, err := New(tokenString, signatureAlgorithms)
	assert.NoError(t, err)
	assert.NotNil(t, webToken)
	assert.Equal(t, issuer, webToken.Issuer)
}

func TestNewAndFail(t *testing.T) {
	tokenString := "just a string"
	_, err := New(tokenString, signatureAlgorithms)
	assert.Error(t, err)
}

func TestNew_DeserializationError(t *testing.T) {
	// Create a valid JWT header and signature, but with a payload that is not valid JSON
	// or cannot be unmarshaled into the expected struct.
	// We'll use jose to construct a token with a payload that is not a JSON object.
	invalidPayload := "not-a-json-object"
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: joseTestKey}, nil)
	assert.NoError(t, err)

	object, err := signer.Sign([]byte(invalidPayload))
	assert.NoError(t, err)

	tokenString, err := object.CompactSerialize()
	assert.NoError(t, err)

	_, err = New(tokenString, signatureAlgorithms)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to deserialize claims")
}
