package graphql

import (
	"mime"
	"net/http"
	"strings"
	"time"

	modulusHttp "github.com/go-modulus/modulus/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func NewHandlerRoute(handler *handler.Server, config Config) (modulusHttp.RouteProvider, modulusHttp.RouteProvider) {
	return modulusHttp.ProvideRawRoute(http.MethodGet, config.Path, handler),
		modulusHttp.ProvideRawRoute(http.MethodPost, config.Path, wrapStreamingHandler(handler))
}

func wrapStreamingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldDisableStreamingTimeouts(r) {
			rc := http.NewResponseController(w)
			_ = rc.SetWriteDeadline(time.Time{})
		}

		next.ServeHTTP(w, r)
	})
}

func shouldDisableStreamingTimeouts(r *http.Request) bool {
	if !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return false
	}
	return r.Method == http.MethodPost && mediaType == "application/json"
}

func NewPlaygroundHandlerRoute(config Config) modulusHttp.RouteProvider {
	if config.Playground.Enabled {
		return modulusHttp.ProvideRawRoute(
			http.MethodGet,
			config.Playground.Path,
			playground.Handler("Graphql Playground", config.Path),
		)
	}

	return modulusHttp.RouteProvider{}
}
