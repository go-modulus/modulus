package locales

import (
	"embed"
	"github.com/go-modulus/modulus/translation"
	"github.com/vorlif/spreak"
)

//go:embed *
var FS embed.FS
var Domain = spreak.NoDomain

func ProvideLocalesFs() interface{} {
	return translation.ProvideLocalesFs(Domain, FS)
}
