package graphql

import (
	"blog/internal/graphql/generated"
	"blog/internal/graphql/resolver"
	"github.com/go-modulus/modulus/module"

	"github.com/99designs/gqlgen/graphql"
)

func NewSchema(r *resolver.Resolver) graphql.ExecutableSchema {
	c := generated.Config{Resolvers: r, Directives: r.GetDirectives()}

	return generated.NewExecutableSchema(c)
}

func NewModule() *module.Module {
	return module.NewModule(
		"graphql",
	).AddProviders(
		resolver.NewResolver,
		NewSchema,
	)
}
