package translation

import (
	"github.com/vorlif/spreak"
	"github.com/vorlif/spreak/localize"
	"golang.org/x/text/language"
)

type Translator struct {
	matcher language.Matcher
	bundle  *spreak.Bundle
}

func NewTranslator(
	bundle *spreak.Bundle,
	matcher language.Matcher,
) *Translator {
	return &Translator{
		bundle:  bundle,
		matcher: matcher,
	}
}
func (t *Translator) NewLocalizer(locale string) *spreak.Localizer {
	tag := t.GetSupportedLocale(locale)
	// Create a Localizer to select the language to translate.
	return spreak.NewLocalizer(t.bundle, tag)
}

func (t *Translator) GetSupportedLocale(locale string) language.Tag {
	tag, _ := language.MatchStrings(t.matcher, locale)

	return tag
}

// E marks a string for extracting into the locales files.
func E(txt localize.Singular) string {
	// It is a hack to mark the string for extracting to the translation file
	_ = localize.MsgID(txt)
	return txt
}
