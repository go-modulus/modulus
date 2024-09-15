package translation

import (
	"net/http"
)

type Middleware struct {
	translator *Translator
}

func NewMiddleware(
	translator *Translator,
) *Middleware {
	return &Middleware{translator: translator}
}

func (a *Middleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			locale := r.Header.Get("Accept-Language")
			ctx = WithLocale(ctx, locale)
			ctx = WithTranslator(ctx, a.translator)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		},
	)
}
