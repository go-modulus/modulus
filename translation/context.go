package translation

import (
	"context"
	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
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

func WithLocalizer(ctx context.Context, localizer *spreak.Localizer) context.Context {
	return context.WithValue(ctx, contextKey("Localizer"), localizer)
}

func GetLocalizer(ctx context.Context) (*spreak.Localizer, error) {
	if value := ctx.Value(contextKey("Localizer")); value != nil {
		return value.(*spreak.Localizer), nil
	}

	// default localizer that uses English language and returns the same text
	opts := []spreak.BundleOption{
		spreak.WithSourceLanguage(language.English),
		spreak.WithLanguage(language.English),
	}

	bundle, err := spreak.NewBundle(
		opts...,
	)
	if err != nil {
		// If we cannot create a default localizer, return nil
		return nil, err
	}
	localizer := spreak.NewLocalizer(bundle, language.English)
	return localizer, nil
}
