package auth

import (
	"braces.dev/errtrace"
	"context"
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

// Middleware is a middleware that authenticates the request and adds the performer to the context.
// It also adds the refresh token to the context.
// It writes the new refresh token to the response.
// @DEPRECATED: Use AddRefreshToken and AddPerformer instead.
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

func (a *Middleware) authenticate(ctx context.Context, authorizationHeader string) (Performer, error) {
	if authorizationHeader == "" {
		return Performer{}, nil
	}
	token, err := a.parseAccessToken(authorizationHeader)
	if err != nil {
		return Performer{}, err
	}
	performer, err := a.authenticator.Authenticate(ctx, token)
	if err != nil {
		if errors.Is(err, ErrTokenIsRevoked) || errors.Is(err, ErrTokenIsExpired) {
			return Performer{}, ErrUnauthenticated
		}
		return Performer{}, errtrace.Wrap(err)
	}
	if performer.ID == uuid.Nil {
		return Performer{}, ErrInvalidToken
	}
	return performer, nil
}

func (a *Middleware) HttpMiddleware() func(http.Handler) http.Handler {
	return errhttp.WrapMiddleware(a.errorPipeline, a.Middleware)
}

func (a *Middleware) AddPerformer(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authorization := r.Header.Get("Authorization")
			performer, err := a.authenticator.Authenticate(ctx, authorization)
			if err != nil {
				ctx = WithError(ctx, err)
			} else {
				ctx = WithPerformer(ctx, performer)
				ctx = logger.AddTags(ctx, "performerId", performer.ID.String())
			}

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func (a *Middleware) AddRefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var refreshToken string
			c, err := r.Cookie(a.config.CookieName)
			if err == nil {
				refreshToken = c.Value
			}
			ctx := WithRefreshToken(r.Context(), refreshToken)

			next.ServeHTTP(
				&refreshTokenResponseWriter{ResponseWriter: w, ctx: ctx, config: a.config.RefreshTokenConfig},
				r.WithContext(ctx),
			)
		},
	)
}

func (a *Middleware) parseAccessToken(token string) (string, error) {
	matches := authRegexp.FindStringSubmatch(token)
	if len(matches) > 2 {
		return matches[2], nil
	}
	return "", ErrInvalidToken
}
