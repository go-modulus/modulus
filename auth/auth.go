package auth

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/errors/errbuilder"
	"github.com/go-modulus/modulus/errors/errhttp"
	"github.com/go-modulus/modulus/logger"
	"github.com/gofrs/uuid"
	"github.com/sethvargo/go-envconfig"
	"net/http"
	"regexp"
	"time"
)

const TagUnauthenticated = "unauthenticated"
const TagUnauthorized = "unauthorized"

var ErrInvalidToken = errbuilder.New("invalid access token").
	WithHint("Please provide a valid access token").
	Build()
var ErrUnauthenticated = errbuilder.New("unauthenticated").
	WithHint("Please authenticate to get access").
	WithTags(TagUnauthenticated).
	Build()
var ErrUnauthorized = errbuilder.New("unauthorized").
	WithHint("You are not authorized to access this resource").
	WithTags(TagUnauthorized).
	Build()

type Performer struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Roles     []string
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (Performer, error)
}

type contextKey string

var performerKey = contextKey("Performer")

func WithPerformer(ctx context.Context, performer Performer) context.Context {
	return context.WithValue(ctx, performerKey, performer)
}

func GetPerformer(ctx context.Context) Performer {
	if value := ctx.Value(performerKey); value != nil {
		performer, ok := value.(Performer)
		if ok {
			return performer
		}
	}
	return Performer{}
}

func GetPerformerID(ctx context.Context) uuid.UUID {
	if value := ctx.Value(performerKey); value != nil {
		performer, ok := value.(Performer)
		if ok {
			return performer.ID
		}
	}
	return uuid.Nil
}

var refreshTokenKey = contextKey("RefreshToken")

type refreshTokenContainer struct {
	value string
	wrote bool
}

func WithRefreshToken(ctx context.Context, refreshToken string) context.Context {
	return context.WithValue(ctx, refreshTokenKey, &refreshTokenContainer{value: refreshToken})
}

func SendRefreshToken(ctx context.Context, token string) {
	if value := ctx.Value(refreshTokenKey); value != nil {
		refreshToken, ok := value.(*refreshTokenContainer)
		if ok {
			refreshToken.value = token
			refreshToken.wrote = true
		}
	}
}

func RemoveRefreshToken(ctx context.Context) {
	SendRefreshToken(ctx, "")
}

func GetRefreshToken(ctx context.Context) string {
	if value := ctx.Value(refreshTokenKey); value != nil {
		refreshToken, ok := value.(*refreshTokenContainer)
		if ok {
			return refreshToken.value
		}
	}
	return ""
}

type refreshTokenResponseWriter struct {
	http.ResponseWriter
	ctx     context.Context
	written bool
	config  RefreshTokenConfig
}

func (rw *refreshTokenResponseWriter) Write(b []byte) (int, error) {
	rw.writeCookie()
	return rw.ResponseWriter.Write(b)
}

func (rw *refreshTokenResponseWriter) WriteHeader(code int) {
	rw.writeCookie()
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *refreshTokenResponseWriter) writeCookie() {
	if rw.written {
		return
	}
	rw.written = true

	value := rw.ctx.Value(refreshTokenKey)
	if value == nil {
		return
	}
	rfc, ok := value.(*refreshTokenContainer)
	if !ok {
		return
	}

	refreshToken := rfc.value
	if !rfc.wrote {
		return
	}

	cookie := http.Cookie{
		Name:     rw.config.CookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		Domain:   rw.config.CookieDomain,
	}
	if rw.config.CookieSecure {
		cookie.Secure = true
		cookie.SameSite = http.SameSiteNoneMode
	}
	if refreshToken == "" {
		cookie.Expires = time.Unix(1, 0)
		cookie.MaxAge = -1
	} else {
		cookie.Value = refreshToken
		cookie.Expires = time.Now().Add(rw.config.TTL)
	}

	rw.ResponseWriter.Header().Add("Set-Cookie", cookie.String())
	rw.ResponseWriter.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
}

func (rw *refreshTokenResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

type RefreshTokenConfig struct {
	CookieName   string        `env:"REFRESH_TOKEN_COOKIE_NAME, default=art"`
	CookieDomain string        `env:"REFRESH_TOKEN_COOKIE_DOMAIN, default=localhost"`
	CookieSecure bool          `env:"REFRESH_TOKEN_COOKIE_SECURE, default=false"`
	TTL          time.Duration `env:"AUTH_REFRESH_TOKEN_TTL, default=8760h"`
}

type MiddlewareConfig struct {
	RefreshTokenConfig
}

func NewMiddlewareConfig() (*MiddlewareConfig, error) {
	config := MiddlewareConfig{}
	return &config, envconfig.Process(context.Background(), &config)
}

var authRegexp = regexp.MustCompile(`(Bearer[ ]+)([^,\n$ ]+)`)

type Middleware struct {
	authenticator Authenticator
	config        *MiddlewareConfig
}

func NewMiddleware(
	authenticator Authenticator,
	config *MiddlewareConfig,
) *Middleware {
	return &Middleware{authenticator: authenticator, config: config}
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

func (a *Middleware) parseAccessToken(token string) (string, error) {
	matches := authRegexp.FindStringSubmatch(token)
	if len(matches) > 2 {
		return matches[2], nil
	}
	return "", ErrInvalidToken
}
