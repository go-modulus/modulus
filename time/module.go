package time

import (
	"github.com/jonboulle/clockwork"
	"go.uber.org/fx"
	_ "golang.org/x/text/message"
)

func NewModule() fx.Option {
	return fx.Module(
		"time",
		fx.Provide(
			clockwork.NewRealClock,
		),
	)
}
