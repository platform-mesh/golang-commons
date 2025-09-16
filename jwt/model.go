package jwt

import (
	"fmt"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

type IssuerAttributes struct {
	Issuer  string `json:"iss"`
	Subject string `json:"sub"`
}

// UserAttributes contains the list of attributes sent to the application by the OIDC Provider
type UserAttributes struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// ParsedAttributes exposes the claims which require of treatment on our side due to incompatibilities between IAS Applications
type ParsedAttributes struct {
	Audiences []string `json:"aud"`
	Mail      string   `json:"mail,omitempty"`
}

// WebToken contains a deserialized id_token sent to the application by the IAS Tenant
type WebToken struct {
	IssuerAttributes
	UserAttributes
	ParsedAttributes
}

// New retrieves a new WebToken from an id_token string provided by OpenID communication
// When not able to parse or deserialize the requested claims, it will return an error
// JWT Claims are parsed without verification, ensure properer JWT verification before calling this function, eg. with istio
func New(idToken string, signatureAlgorithms []jose.SignatureAlgorithm) (webToken WebToken, err error) {
	token, parseErr := jwt.ParseSigned(idToken, signatureAlgorithms)
	if parseErr != nil {
		err = fmt.Errorf("unable to parse id_token: %w", parseErr)
		return
	}

	rawToken := new(rawWebToken)
	desErr := token.UnsafeClaimsWithoutVerification(&rawToken)
	if desErr != nil {
		err = fmt.Errorf("unable to deserialize claims: %w", desErr)
		return
	}

	webToken.UserAttributes = rawToken.UserAttributes
	webToken.IssuerAttributes = rawToken.IssuerAttributes
	webToken.Audiences = rawToken.getAudiences()
	webToken.Mail = rawToken.getMail()
	webToken.FirstName = rawToken.getFirstName()
	webToken.LastName = rawToken.getLastName()

	return
}
