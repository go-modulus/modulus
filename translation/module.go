package translation

import (
	"errors"
	"io/fs"

	"github.com/go-modulus/modulus/module"
	"github.com/vorlif/spreak"
	"go.uber.org/fx"
	"golang.org/x/text/language"
)

type ModuleConfig struct {
	Locales []string `env:"TRANSLATION_LOCALES, default=en-US,uk-UA" comment:"List of supported locales for translation. Example: TRANSLATION_LOCALES=en-US,uk-UA"`
}

type BundleParams struct {
	fx.In

	Fs     []LocalesFolder `group:"translation.locales-fs"`
	Config ModuleConfig
}
type LocalesFolder struct {
	Domain string `json:"domain"`
	Fs     fs.FS  `json:"fs"`
}

func NewModule() *module.Module {
	return module.NewModule("translation").
		AddProviders(
			func(cfg ModuleConfig) language.Matcher {
				tags := make([]language.Tag, 0, len(cfg.Locales))
				for _, locale := range cfg.Locales {
					tags = append(tags, language.MustParse(locale))
				}
				return language.NewMatcher(
					tags,
				)
			},
			func(params BundleParams) (*spreak.Bundle, error) {
				cfg := params.Config
				tags := make([]interface{}, 0, len(cfg.Locales))
				for _, locale := range cfg.Locales {
					tag, err := language.Parse(locale)
					if err != nil {
						return nil, errors.New(locale + " is not a valid language")
					}
					tags = append(tags, tag)
				}
				opts := []spreak.BundleOption{
					// Set the language used in the program code/templates
					spreak.WithSourceLanguage(language.English),
					// Specify the languages you want to load
					spreak.WithLanguage(tags...),
				}

				//// Merge all filesystems into a single in-memory filesystem
				//// This process: 1) Goes through all *.po files in each FS
				//// 2) Merges content of files with same name
				//// 3) Saves merged files in new in-memory FS
				//if len(params.Fs) > 0 {
				//	mergedFS, err := mergePoFilesystems(params.Fs)
				//	if err != nil {
				//		return nil, err
				//	}
				//	opts = append(opts, spreak.WithDomainFs(spreak.NoDomain, mergedFS))
				//}

				if len(params.Fs) > 0 {
					domainsMap := make(map[string]fs.FS)
					for _, f := range params.Fs {
						if f.Fs == nil {
							continue // Skip if FS is nil
						}
						domainsMap[f.Domain] = f.Fs
					}
					if len(domainsMap) > 0 {
						for domain, f := range domainsMap {
							if len(domain) > 0 {
								opts = append(opts, spreak.WithDomainFs(domain, f))
							} else {
								opts = append(opts, spreak.WithDomainFs(spreak.NoDomain, f))
							}
						}
					}
				}
				return spreak.NewBundle(
					opts...,
				)
			},
			NewTranslator,
			NewMiddleware,
		).AddInvokes(
		func(c ModuleConfig) {

		},
	).InitConfig(ModuleConfig{})

}

func NewManifestModule() module.ManifestModule {
	graphqlModule := module.NewManifestModule(
		NewModule(),
		"github.com/go-modulus/modulus/translation",
		"Provides translation services using Spreak library",
		"1.0.0",
	)
	graphqlModule.Install.AppendFiles(
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/translation/install/translation.mk",
			DestFile:  "mk/translation.mk",
		},
		module.InstalledFile{
			SourceUrl: "https://raw.githubusercontent.com/go-modulus/modulus/refs/heads/main/translation/install/locales.go",
			DestFile:  "locales/locales.go",
		},
	)

	return graphqlModule
}

func ProvideLocalesFs(
	domain string,
	localesFS fs.FS,
) interface{} {
	return fx.Annotate(
		func() LocalesFolder {
			return LocalesFolder{
				Domain: domain,
				Fs:     localesFS,
			}
		},
		fx.ResultTags(`group:"translation.locales-fs"`),
	)
}
