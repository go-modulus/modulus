package translation

import (
	"errors"
	"github.com/go-modulus/modulus/module"
	"github.com/spf13/afero"
	"github.com/vorlif/spreak"
	"go.uber.org/fx"
	"golang.org/x/text/language"
	"io/fs"
)

type ModuleConfig struct {
	Locales     []string `env:"TRANSLATION_LOCALES, default=en-US,uk-UA" comment:"List of supported locales for translation. Example: TRANSLATION_LOCALES=en-US,uk-UA"`
	LocalesPath string   `env:"TRANSLATION_LOCALES_PATH"`
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
	return module.NewModule("modulus/translation").
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

func mergePoFilesystems(fsList []fs.FS) (fs.FS, error) {
	if len(fsList) == 0 {
		return nil, errors.New("no filesystems provided to merge")
	}
	mem := afero.NewMemMapFs()

	filesContent := make(map[string]string)
	for _, f := range fsList {
		err := fs.WalkDir(
			f, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !d.Type().IsRegular() {
					return nil // Skip directories and non-regular files
				}
				content, err := fs.ReadFile(f, path)
				if err != nil {
					return err
				}
				// Store content in a map to merge later
				filesContent[path] += string(content) // Concatenate content if same file exists in multiple FS
				return nil
			},
		)
		if err != nil {
			return nil, err
		}
	}

	for path, content := range filesContent {
		// Write merged content to the in-memory filesystem
		if err := afero.WriteFile(mem, path, []byte(content), 0644); err != nil {
			return nil, err
		}
	}
	return afero.NewIOFS(mem), nil
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
