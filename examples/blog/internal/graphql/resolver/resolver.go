package resolver

import (
	"blog/internal/auth/graphql"
	blogGraphql "blog/internal/blog/graphql"
	"blog/internal/graphql/generated"
	userGraphql "blog/internal/user/graphql"
	userDataloader "blog/internal/user/storage/dataloader"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// Place all dependencies here
	blogResolver      *blogGraphql.Resolver
	userResolver      *userGraphql.Resolver
	userLoaderFactory *userDataloader.LoaderFactory
}

func NewResolver(
	blogResolver *blogGraphql.Resolver,
	userResolver *userGraphql.Resolver,
	userLoaderFactory *userDataloader.LoaderFactory,
) *Resolver {
	return &Resolver{
		blogResolver:      blogResolver,
		userResolver:      userResolver,
		userLoaderFactory: userLoaderFactory,
	}
}

func (r Resolver) GetDirectives() generated.DirectiveRoot {
	return generated.DirectiveRoot{
		AuthGuard: graphql.AuthGuard,
	}
}
