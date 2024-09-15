package translation

import (
	"go.uber.org/fx"
	"golang.org/x/text/language"
)

type ModuleConfig struct {
	Locales []string `env:"TRANSLATION_LOCALES"`
}

func NewModule(cfg ModuleConfig) fx.Option {
	return fx.Module(
		"modulus/translation",
		fx.Provide(
			func() language.Matcher {
				tags := make([]language.Tag, 0, len(cfg.Locales))
				for _, locale := range cfg.Locales {
					tags = append(tags, language.MustParse(locale))
				}
				return language.NewMatcher(
					tags,
				)
			},
			NewTranslator,
			NewMiddleware,
		),
	)
}
