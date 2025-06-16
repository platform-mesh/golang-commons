package keys

import "github.com/platform-mesh/golang-commons/jwt"

type ContextKey string

const (
	RequestIdCtxKey     = ContextKey("request-id")
	LoggerCtxKey        = ContextKey("logger")
	ConfigCtxKey        = ContextKey("config")
	SentryTagsCtxKey    = ContextKey("sentryTags")
	TechnicalUserCtxKey = ContextKey("technicalUser")
	SpiffeCtxKey        = ContextKey(jwt.SpiffeCtxKey)
	TenantIdCtxKey      = ContextKey(jwt.TenantIdCtxKey)
	AuthHeaderCtxKey    = ContextKey(jwt.AuthHeaderCtxKey)
	WebTokenCtxKey      = ContextKey(jwt.WebTokenCtxKey)
	UserIDCtxKey        = ContextKey("userId")
)
