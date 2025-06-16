package sentry

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	pmcontext "github.com/platform-mesh/golang-commons/context"
	"github.com/platform-mesh/golang-commons/logger"
)

// GraphQLErrorPresenter returns a function that can be used as GraphQL error presenter
func GraphQLErrorPresenter(skipTenants ...string) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)
		if err == nil {
			return nil
		}

		if !IsSentryError(e) {
			l := logger.LoadLoggerFromContext(ctx)

			spiffe, err2 := pmcontext.GetSpiffeFromContext(ctx)
			isTechnicalIssuer := pmcontext.GetIsTechnicalIssuerFromContext(ctx)
			webToken, err3 := pmcontext.GetWebTokenFromContext(ctx)

			event := l.Debug().Err(err)
			if err2 == nil {
				event = event.Str("spiffe", spiffe)
			}
			if err3 == nil {
				event = event.Interface("webToken.Subject", webToken.Subject)
			}
			event = event.Bool("isTechnicalIssuer", isTechnicalIssuer)
			event.Msg("Error not sent to Sentry")
			return err
		}

		if !pmcontext.HasTenantInContext(ctx) {
			captureErrorForContext(ctx, err, "")
		}

		tenantID, _ := pmcontext.GetTenantFromContext(ctx)

		// return without sending to Sentry if tenant should be skipped
		for _, tenant := range skipTenants {
			if tenant == tenantID {
				l := logger.LoadLoggerFromContext(ctx)
				l.Debug().Err(err).Msg("Error not sent to Sentry for skipped tenant")
				return err
			}
		}

		captureErrorForContext(ctx, err, tenantID)

		return err
	}
}

// GraphQLRecover returns a function that can be used as GraphQL error presenter
func GraphQLRecover(log *logger.Logger) graphql.RecoverFunc {
	return func(ctx context.Context, err interface{}) (userMessage error) {
		log.Error().Interface("stack", debug.Stack()).Msgf("GraphQL panic: %v", err)

		tenantID, ctxErr := pmcontext.GetTenantFromContext(ctx)
		if ctxErr != nil {
			captureErrorForContext(ctx, ctxErr, "")
		}

		captureErrorForContext(ctx, fmt.Errorf("GraphQL panic: %v", err), tenantID)

		return gqlerror.Errorf("internal server error: %v", err)
	}
}

// captureErrorForContext sends the error to Sentry and adds tags and extras from context if possible
func captureErrorForContext(ctx context.Context, err error, tenantID string) {
	extras := Extras{}
	tags := Tags{}
	if graphql.HasOperationContext(ctx) {
		oc := graphql.GetOperationContext(ctx)
		if oc != nil {
			extras.Add("operation", oc.Operation.Operation)
			extras.Add("variables", oc.Variables)
			extras.Add("query", oc.RawQuery)
		}
	}

	path := graphql.GetPath(ctx)
	if path != nil {
		tags.Add("path", path.String())
	}

	if tenantID != "" {
		tags.Add("tenantID", tenantID)
	}

	CaptureError(err, tags, extras)
}
