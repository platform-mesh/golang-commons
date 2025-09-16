package jwt

type rawClaims struct {
	RawAudiences  interface{} `json:"aud"` // RawAudiences could be a []string or string depending on the serialization in IdP site
	RawEmail      string      `json:"email,omitempty"`
	RawMail       string      `json:"mail,omitempty"`
	RawGivenName  string      `json:"given_name,omitempty"`
	RawFamilyName string      `json:"family_name,omitempty"`
}

type rawWebToken struct {
	rawClaims
	IssuerAttributes
	UserAttributes
}

func (r rawWebToken) getMail() (mail string) {
	mail = r.RawMail
	if mail == "" {
		mail = r.RawEmail
	}
	return
}

func (r rawWebToken) getLastName() (lastName string) {
	lastName = r.LastName
	if lastName == "" {
		lastName = r.RawFamilyName
	}
	return
}

func (r rawWebToken) getFirstName() (firstName string) {
	firstName = r.FirstName
	if firstName == "" {
		firstName = r.RawGivenName
	}
	return
}

func (r rawWebToken) getAudiences() (audiences []string) {
	switch audienceList := r.RawAudiences.(type) {
	case string:
		audiences = []string{audienceList}
	case []interface{}:
		for _, val := range audienceList {
			aud, ok := val.(string)
			if ok {
				audiences = append(audiences, aud)
			}
		}
	}

	return
}
