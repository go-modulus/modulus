package locales

import (
	"embed"
	"github.com/go-modulus/modulus/translation"
)

//go:embed *
var FS embed.FS
var Domain = "auth"

func ProvideLocalesFs() interface{} {
	return translation.ProvideLocalesFs(Domain, FS)
}
