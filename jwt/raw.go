package jwt

type rawClaims struct {
	RawAudiences interface{} `json:"aud"` // RawAudiences could be a []string or string depending on the serialization in IdP site
	RawEmail     string      `json:"email,omitempty"`
	RawMail      string      `json:"mail,omitempty"`
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
