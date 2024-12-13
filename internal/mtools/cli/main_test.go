package cli_test

import (
	"github.com/go-modulus/modulus/internal/mtools"
	"github.com/go-modulus/modulus/internal/mtools/cli"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"testing"
)

var (
	addModule    *cli.AddModule
	createModule *cli.CreateModule
)

func TestMain(m *testing.M) {
	currentModule := mtools.NewModule()
	test.TestMain(
		m,
		currentModule.BuildFx(),
		fx.Populate(
			&addModule,
			&createModule,
		),
	)

}
