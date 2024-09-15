package translation

import (
	"context"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type contextKey string

func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, contextKey("Locale"), locale)
}

func GetLocale(ctx context.Context) string {
	locale := ""
	if value := ctx.Value(contextKey("Locale")); value != nil {
		strVal, ok := value.(string)
		if ok {
			locale = strVal
		}
	}
	return locale
}

func WithTranslator(ctx context.Context, translator *Translator) context.Context {
	return context.WithValue(ctx, contextKey("Translator"), translator)
}

func GetTranslator(ctx context.Context) *Translator {
	if value := ctx.Value(contextKey("Translator")); value != nil {
		return value.(*Translator)
	}

	return nil
}

func GetPrinter(ctx context.Context) *message.Printer {
	locale := GetLocale(ctx)
	t := GetTranslator(ctx)
	if t == nil {
		return message.NewPrinter(language.English)
	}
	tag := t.GetSupportedLocale(locale)

	return message.NewPrinter(tag)

}
