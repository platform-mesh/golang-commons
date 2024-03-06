package config

import (
	"context"

	"github.com/openmfp/golang-commons/context/keys"
)

func SetConfigInContext(ctx context.Context, config any) context.Context {
	return context.WithValue(ctx, keys.ConfigCtxKey, config)
}

func LoadConfigFromContext(ctx context.Context) any {
	return ctx.Value(keys.ConfigCtxKey)
}
