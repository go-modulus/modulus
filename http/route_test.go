package http_test

import (
	"github.com/go-modulus/modulus/http"
	"github.com/stretchr/testify/require"
	http2 "net/http"
	"testing"
)

func TestRoute_IsEmpty(t *testing.T) {
	t.Parallel()
	t.Run(
		"IsEmpty for empty path", func(t *testing.T) {
			t.Parallel()
			r := http.ProvideRawRoute(
				"GET",
				"",
				http2.HandlerFunc(
					func(w http2.ResponseWriter, r *http2.Request) {

					},
				),
			)

			require.True(t, r.Route.IsEmpty())
		},
	)

	t.Run(
		"IsEmpty for empty handler", func(t *testing.T) {
			t.Parallel()
			r := http.ProvideRawRoute(
				"GET",
				"/test",
				nil,
			)

			require.True(t, r.Route.IsEmpty())
		},
	)

	t.Run(
		"IsEmpty for empty err handler", func(t *testing.T) {
			t.Parallel()
			r := http.ProvideRoute(
				"GET",
				"/test",
				nil,
			)

			require.True(t, r.Route.IsEmpty())
		},
	)

	t.Run(
		"Not empty if err handler is present", func(t *testing.T) {
			r := http.ProvideRoute(
				"GET",
				"/test",
				func(w http2.ResponseWriter, req *http2.Request) error {
					return nil
				},
			)

			require.False(t, r.Route.IsEmpty())
		},
	)

	t.Run(
		"not emptty if handler is present", func(t *testing.T) {
			t.Parallel()
			r := http.ProvideRawRoute(
				"GET",
				"/test",
				http2.HandlerFunc(
					func(w http2.ResponseWriter, r *http2.Request) {

					},
				),
			)

			require.False(t, r.Route.IsEmpty())
		},
	)

}
