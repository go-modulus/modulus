package graphql_test

import (
	"blog/internal/blog"
	"blog/internal/blog/graphql"
	"blog/internal/blog/storage/fixture"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"testing"
)

var (
	resolver *graphql.Resolver
	fixtures *fixture.Factory
)

func TestMain(m *testing.M) {
	test.LoadEnv("../../..")
	mod := blog.NewModule().
		AddProviders(fixture.NewFactory)

	test.TestMain(
		m,
		module.BuildFx(mod),
		fx.Populate(
			&resolver,
			&fixtures,
		),
	)
}
