package graphql

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/go-modulus/modulus/auth"
	"github.com/gofrs/uuid"
)

const DefaultUserRole = "user"

func AuthGuard(
	ctx context.Context,
	_ any,
	next graphql.Resolver,
	allowedRoles []string,
) (res any, err error) {
	performer := auth.GetPerformer(ctx)
	if performer.ID == uuid.Nil {
		return nil, auth.ErrUnauthenticated
	}

	if len(allowedRoles) == 0 {
		return next(ctx)
	}

	userRolesMap := make(map[string]struct{}, len(performer.Roles))
	for _, role := range performer.Roles {
		userRolesMap[role] = struct{}{}
	}

	for _, role := range allowedRoles {
		if _, ok := userRolesMap[role]; ok {
			return next(ctx)
		}
	}

	return nil, auth.ErrUnauthorized
}
