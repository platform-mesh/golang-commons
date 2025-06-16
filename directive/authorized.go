package directive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	pmcontext "github.com/platform-mesh/golang-commons/context"
	"github.com/platform-mesh/golang-commons/fga/helpers"
	"github.com/platform-mesh/golang-commons/logger"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/metadata"
)

func extractNestedKeyFromArgs(args map[string]any, paramName string) (string, error) {
	o, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	var normalizedArgs map[string]any
	err = json.Unmarshal(o, &normalizedArgs)
	if err != nil {
		return "", err
	}

	var paramValue string
	parts := strings.Split(paramName, ".")
	for i, key := range parts {
		val, ok := normalizedArgs[key]
		if !ok {
			return "", fmt.Errorf("unable to extract param from request for given paramName %q", paramName)
		}

		if i == len(strings.Split(paramName, "."))-1 {
			paramValue, ok = val.(string)
			if !ok || paramValue == "" {
				return "", fmt.Errorf("unable to extract param from request for given paramName %q, param is of wrong type", paramName)
			}

			return paramValue, nil
		}

		normalizedArgs, ok = val.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("unable to extract param from request for given paramName %q, param is of wrong type", paramName)
		}
	}

	return paramValue, nil
}

func Authorized(openfgaClient openfgav1.OpenFGAServiceClient, log *logger.Logger) func(context.Context, interface{}, graphql.Resolver, string, *string, *string, string) (interface{}, error) {
	if !directiveConfiguration.DirectivesAuthorizationEnabled {
		return func(ctx context.Context, obj interface{}, next graphql.Resolver, relation string, entityType *string, entityTypeParamName *string, entityParamName string) (interface{}, error) {
			return next(ctx)
		}
	}

	return func(ctx context.Context, obj interface{}, next graphql.Resolver, relation string, entityType *string, entityTypeParamName *string, entityParamName string) (interface{}, error) {

		if openfgaClient == nil {
			return nil, errors.New("OpenFGAServiceClient is nil. Cannot process request")
		}

		ctx, err := setTenantToContextForTechnicalUsers(ctx, log)
		if err != nil {
			return nil, err
		}

		token, err := pmcontext.GetAuthHeaderFromContext(ctx)
		hasToken := err == nil

		if hasToken {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
		}

		fctx := graphql.GetFieldContext(ctx)

		entityID, err := extractNestedKeyFromArgs(fctx.Args, entityParamName)
		if err != nil {
			return nil, err
		}

		tenantID, err := pmcontext.GetTenantFromContext(ctx)
		if err != nil {
			return nil, err
		}

		evaluatedEntityType := ""
		if entityTypeParamName != nil {
			evaluatedEntityType, err = extractNestedKeyFromArgs(fctx.Args, *entityTypeParamName)
			if err != nil {
				return nil, err
			}
		} else if entityType != nil {
			evaluatedEntityType = *entityType
		}

		if evaluatedEntityType == "" {
			return nil, fmt.Errorf("make sure to either provide entityType or entityTypeParamName")
		}

		storeID, err := helpers.GetStoreIDForTenant(ctx, openfgaClient, tenantID)
		if err != nil {
			return nil, err
		}
		modelID, err := helpers.GetModelIDForTenant(ctx, openfgaClient, tenantID)
		if err != nil {
			return nil, err
		}

		var userID string
		if hasToken {
			user, err := pmcontext.GetWebTokenFromContext(ctx)
			if err != nil {
				return nil, err
			}
			userID = user.Subject
		} else {
			spiffe, err := pmcontext.GetSpiffeFromContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("authorized was invoked without a user token or a spiffe header")
			}
			userID = strings.TrimPrefix(spiffe, "spiffe://")
			log.Trace().Str("user", userID).Msg("using spiffe user in authorized directive")
		}

		req := &openfgav1.CheckRequest{
			StoreId:              storeID,
			AuthorizationModelId: modelID,
			TupleKey: &openfgav1.CheckRequestTupleKey{
				User:     fmt.Sprintf("user:%s", helpers.SanitizeUserID(userID)),
				Relation: relation,
				Object:   fmt.Sprintf("%s:%s", evaluatedEntityType, entityID),
			},
		}

		res, err := openfgaClient.Check(ctx, req)
		if err != nil {
			log.Error().Err(err).Str("user", req.TupleKey.User).Msg("authorization check failed")
			return nil, err
		}

		if !res.Allowed {
			log.Warn().Bool("allowed", res.Allowed).Any("req", req).Msg("not allowed")
			return nil, gqlerror.Errorf("unauthorized")
		}

		return next(ctx)
	}
}
