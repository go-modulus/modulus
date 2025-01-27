package resolver

import (
	"blog/internal/graphql/generated"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// Place all dependencies here
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r Resolver) GetDirectives() generated.DirectiveRoot {
	return generated.DirectiveRoot{}
}
