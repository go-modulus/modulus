//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/rakyll/gotest"
	_ "github.com/vektra/mockery/v2"

	_ "github.com/go-modulus/modulus"

	_ "github.com/go-modulus/modulus/db/migrator"
	_ "github.com/go-modulus/modulus/db/pgx"
	_ "github.com/go-modulus/modulus/graphql"
	_ "github.com/go-modulus/modulus/http"
	_ "github.com/go-modulus/modulus/logger"
)
