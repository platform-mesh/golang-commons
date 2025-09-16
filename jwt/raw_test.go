package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAudiences(t *testing.T) {
	rawAudiences := []interface{}{
		"audience1",
		"audience2",
		1812, // wrong audience
	}

	token := rawWebToken{
		rawClaims: rawClaims{
			RawAudiences: rawAudiences,
		},
	}

	parsedAudiences := token.getAudiences()

	assert.Contains(t, parsedAudiences, "audience1")
	assert.Contains(t, parsedAudiences, "audience2")
	assert.NotContains(t, parsedAudiences, 1812)
}
func TestParseAudiencesString(t *testing.T) {
	token := rawWebToken{
		rawClaims: rawClaims{
			RawAudiences: "audience1",
		},
	}

	parsedAudiences := token.getAudiences()

	assert.Contains(t, parsedAudiences, "audience1")
	assert.Equal(t, len(parsedAudiences), 1)
}

func TestGetFirstName_PreferFirstName(t *testing.T) {
	token := rawWebToken{
		UserAttributes: UserAttributes{
			FirstName: "John",
		},
		rawClaims: rawClaims{
			RawGivenName: "Jonathan",
		},
	}

	firstName := token.getFirstName()

	assert.Equal(t, "John", firstName)
}

func TestGetFirstName_FallbackToRawGivenName(t *testing.T) {
	token := rawWebToken{
		UserAttributes: UserAttributes{
			FirstName: "",
		},
		rawClaims: rawClaims{
			RawGivenName: "Jonathan",
		},
	}

	firstName := token.getFirstName()

	assert.Equal(t, "Jonathan", firstName)
}

func TestGetFirstName_BothEmpty(t *testing.T) {
	token := rawWebToken{
		UserAttributes: UserAttributes{
			FirstName: "",
		},
		rawClaims: rawClaims{
			RawGivenName: "",
		},
	}

	firstName := token.getFirstName()

	assert.Equal(t, "", firstName)
}

func TestGetLastName_PreferLastName(t *testing.T) {
	token := rawWebToken{
		UserAttributes: UserAttributes{
			LastName: "Doe",
		},
		rawClaims: rawClaims{
			RawFamilyName: "Smith",
		},
	}

	lastName := token.getLastName()

	assert.Equal(t, "Doe", lastName)
}

func TestGetLastName_FallbackToRawFamilyName(t *testing.T) {
	token := rawWebToken{
		UserAttributes: UserAttributes{
			LastName: "",
		},
		rawClaims: rawClaims{
			RawFamilyName: "Smith",
		},
	}

	lastName := token.getLastName()

	assert.Equal(t, "Smith", lastName)
}

func TestGetLastName_BothEmpty(t *testing.T) {
	token := rawWebToken{
		UserAttributes: UserAttributes{
			LastName: "",
		},
		rawClaims: rawClaims{
			RawFamilyName: "",
		},
	}

	lastName := token.getLastName()

	assert.Equal(t, "", lastName)
}
