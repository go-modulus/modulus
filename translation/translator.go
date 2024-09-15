package translation

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Translator struct {
	matcher language.Matcher
}

func NewTranslator(matcher language.Matcher) *Translator {
	return &Translator{
		matcher: matcher,
	}
}
func (t *Translator) NewPrinter(locale string) *message.Printer {
	tag := t.GetSupportedLocale(locale)

	return message.NewPrinter(tag)
}

func (t *Translator) GetSupportedLocale(locale string) language.Tag {
	tag, _ := language.MatchStrings(t.matcher, locale)

	return tag
}
