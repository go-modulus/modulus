package main

import (
	auth2 "blog/internal/auth"
	"blog/internal/blog"
	graphql2 "blog/internal/graphql"
	"blog/internal/user"
	"fmt"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/cli"
	"github.com/go-modulus/modulus/config"
	"github.com/go-modulus/modulus/db/migrator"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/graphql"
	"github.com/go-modulus/modulus/http"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"

	"go.uber.org/fx"
)

func main() {
	fmt.Println("Starting the application...")
	config.LoadDefaultEnv()

	// DO NOT Remove. It will be edited by the `mtools module create` CLI command.
	modules := []*module.Module{
		cli.NewModule().InitConfig(
			cli.ModuleConfig{
				Version: "0.1.0",
				Usage:   "Run project commands",
			},
		),
		pgx.NewModule(),
		migrator.NewModule(),
		http.NewModule(),
		graphql.NewModule(),
		graphql2.NewModule(),
		logger.NewModule(),
		blog.NewModule(),
		user.NewModule(),
		auth.NewModule(),
		auth2.NewModule(),
	}

	invokes := []fx.Option{
		fx.Invoke(cli.Start),
	}

	app := fx.New(
		module.BuildFx(modules...),
		fx.Module("invokes", invokes...),
	)

	app.Run()
}
