package resolver

import (
	"blog/internal/auth/graphql"
	blogGraphql "blog/internal/blog/graphql"
	"blog/internal/graphql/generated"
	userGraphql "blog/internal/user/graphql"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// Place all dependencies here
	blogResolver *blogGraphql.Resolver
	userResolver *userGraphql.Resolver
}

func NewResolver(
	blogResolver *blogGraphql.Resolver,
	userResolver *userGraphql.Resolver,
) *Resolver {
	return &Resolver{
		blogResolver: blogResolver,
		userResolver: userResolver,
	}
}

func (r Resolver) GetDirectives() generated.DirectiveRoot {
	return generated.DirectiveRoot{
		AuthGuard: graphql.AuthGuard,
	}
}
