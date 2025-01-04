package cli_test

import (
	"github.com/go-modulus/modulus/internal/mtools"
	"github.com/go-modulus/modulus/internal/mtools/cli/module"
	module2 "github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"testing"
)

var (
	addModule    *module.Install
	createModule *module.Create
)

func TestMain(m *testing.M) {
	currentModule := mtools.NewModule()
	test.TestMain(
		m,
		module2.BuildFx(currentModule),
		fx.Populate(
			&addModule,
			&createModule,
		),
	)

}
