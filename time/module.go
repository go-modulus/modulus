package time

import (
	"github.com/go-modulus/modulus/module"
	"github.com/jonboulle/clockwork"
	_ "golang.org/x/text/message"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/time").
		AddProviders(
			clockwork.NewRealClock,
		)
}
