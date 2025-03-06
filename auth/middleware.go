package auth

import (
	"braces.dev/errtrace"
	"errors"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/go-modulus/modulus/logger"
	"github.com/gofrs/uuid"
	"net/http"
	"regexp"
)

var authRegexp = regexp.MustCompile(`(Bearer[ ]+)([^,\n$ ]+)`)

type Middleware struct {
	authenticator Authenticator
	config        *MiddlewareConfig
	errorPipeline *errhttp.ErrorPipeline
}

func NewMiddleware(
	authenticator Authenticator,
	config *MiddlewareConfig,
	errorPipeline *errhttp.ErrorPipeline,
) *Middleware {
	return &Middleware{
		authenticator: authenticator,
		config:        config,
		errorPipeline: errorPipeline,
	}
}

func (a *Middleware) Middleware(next http.Handler) errhttp.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var refreshToken string
		c, err := r.Cookie(a.config.CookieName)
		if err == nil {
			refreshToken = c.Value
		}
		ctx := WithRefreshToken(r.Context(), refreshToken)

		authorization := r.Header.Get("Authorization")
		if authorization != "" {
			token, err := a.parseAccessToken(authorization)
			if err != nil {
				return err
			}
			performer, err := a.authenticator.Authenticate(r.Context(), token)
			if err != nil {
				if errors.Is(err, ErrTokenIsRevoked) || errors.Is(err, ErrTokenIsExpired) {
					return ErrUnauthenticated
				}
				return errtrace.Wrap(err)
			}
			if performer.ID == uuid.Nil {
				return ErrInvalidToken
			}

			ctx = WithPerformer(ctx, performer)
			ctx = logger.AddTags(ctx, "performerId", performer.ID.String())
		}

		next.ServeHTTP(
			&refreshTokenResponseWriter{ResponseWriter: w, ctx: ctx, config: a.config.RefreshTokenConfig},
			r.WithContext(ctx),
		)
		return nil
	}
}

func (a *Middleware) HttpMiddleware() func(http.Handler) http.Handler {
	return errhttp.WrapMiddleware(a.errorPipeline, a.Middleware)
}

func (a *Middleware) parseAccessToken(token string) (string, error) {
	matches := authRegexp.FindStringSubmatch(token)
	if len(matches) > 2 {
		return matches[2], nil
	}
	return "", ErrInvalidToken
}
