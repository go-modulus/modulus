package translation

import (
	"github.com/go-modulus/modulus/module"
	"golang.org/x/text/language"
)

type ModuleConfig struct {
	Locales []string `env:"TRANSLATION_LOCALES"`
}

func NewModule(cfg ModuleConfig) *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/translation").
		AddProviders(
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
		)
}
