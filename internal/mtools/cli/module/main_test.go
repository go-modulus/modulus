package module_test

import (
	"github.com/go-modulus/modulus/internal/mtools"
	"github.com/go-modulus/modulus/internal/mtools/cli/module"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"os"
	"testing"
)

var (
	installModule *module.Install
	createModule  *module.Create
)

func TestMain(m *testing.M) {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	test.LoadEnv(currentDir + "/../../../..")
	currentModule := mtools.NewModule()
	test.TestMain(
		m,
		currentModule.BuildFx(),
		fx.Populate(
			&installModule,
			&createModule,
		),
	)
}
