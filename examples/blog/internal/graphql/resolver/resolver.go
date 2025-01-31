package resolver

import (
	blogGraphql "blog/internal/blog/graphql"
	"blog/internal/graphql/generated"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// Place all dependencies here
	blogResolver *blogGraphql.Resolver
}

func NewResolver(
	blogResolver *blogGraphql.Resolver,
) *Resolver {
	return &Resolver{
		blogResolver: blogResolver,
	}
}

func (r Resolver) GetDirectives() generated.DirectiveRoot {
	return generated.DirectiveRoot{}
}
