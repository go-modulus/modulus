package action_test

import (
	"github.com/go-modulus/modulus/internal/mtools"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"testing"
)

var (
	installStorage *action.InstallStorage
)

func TestMain(m *testing.M) {
	currentModule := mtools.NewModule()
	test.TestMain(
		m,
		currentModule.BuildFx(),
		fx.Populate(
			&installStorage,
		),
	)

}
