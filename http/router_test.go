package http_test

import (
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-modulus/modulus/http"
	"github.com/go-modulus/modulus/http/errhttp"
	"github.com/stretchr/testify/require"
)

// rawRequest sends a bare HTTP/1.1 request over TCP and returns everything
// the server wrote back, so hijacked, raw post-upgrade bytes are visible too
// (something a net/http client wouldn't let us observe).
func rawRequest(t *testing.T, addr string, request string) string {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))
	_, err = conn.Write([]byte(request))
	require.NoError(t, err)

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	require.NoError(t, err)

	return string(buf[:n])
}

func TestDefaultRouter_Hijack(t *testing.T) {
	t.Run(
		"upgrades the connection and lets the handler write raw bytes after Hijack", func(t *testing.T) {
			router := http.NewDefaultRouter(&errhttp.ErrorPipeline{}, http.ServeConfig{})
			router.Method(
				"GET", "/ws", nethttp.HandlerFunc(
					func(w nethttp.ResponseWriter, r *nethttp.Request) {
						w.Header().Set("Upgrade", "websocket")
						w.Header().Set("Connection", "Upgrade")
						w.WriteHeader(nethttp.StatusSwitchingProtocols)

						hj, ok := w.(nethttp.Hijacker)
						require.True(t, ok, "ResponseWriter reaching the handler must support Hijack")

						conn, rw, err := hj.Hijack()
						require.NoError(t, err)
						defer conn.Close()

						_, err = rw.Write([]byte("raw-post-hijack-payload"))
						require.NoError(t, err)
						require.NoError(t, rw.Flush())
					},
				),
			)

			server := httptest.NewServer(router)
			defer server.Close()

			addr := server.Listener.Addr().String()
			response := rawRequest(
				t, addr,
				"GET /ws HTTP/1.1\r\nHost: "+addr+"\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n",
			)

			require.Contains(t, response, "101 Switching Protocols")
			require.Contains(t, response, "Upgrade: websocket")
			require.Contains(t, response, "raw-post-hijack-payload")
		},
	)

	t.Run(
		"falls back to a 500 instead of corrupting the response when the underlying writer can't hijack",
		func(t *testing.T) {
			router := http.NewDefaultRouter(&errhttp.ErrorPipeline{}, http.ServeConfig{})
			router.Method(
				"GET", "/ws", nethttp.HandlerFunc(
					func(w nethttp.ResponseWriter, r *nethttp.Request) {
						hj, ok := w.(nethttp.Hijacker)
						require.True(t, ok)
						_, _, err := hj.Hijack()
						if err != nil {
							nethttp.Error(w, "hijack not supported", nethttp.StatusInternalServerError)
						}
					},
				),
			)

			rec := httptest.NewRecorder()
			// httptest.ResponseRecorder does not implement http.Hijacker.
			req := httptest.NewRequest("GET", "/ws", nil)
			router.ServeHTTP(rec, req)

			require.Equal(t, nethttp.StatusInternalServerError, rec.Code)
		},
	)
}

// unwrapOnlyResponseWriter mirrors the shape of real middleware wrappers in
// this codebase (e.g. auth.refreshTokenResponseWriter and the request
// logger's responseWriter): it embeds http.ResponseWriter and exposes Unwrap,
// but does NOT implement http.Hijacker itself.
type unwrapOnlyResponseWriter struct {
	nethttp.ResponseWriter
}

func (w *unwrapOnlyResponseWriter) Unwrap() nethttp.ResponseWriter {
	return w.ResponseWriter
}

func TestDefaultRouter_Hijack_ThroughUnwrapOnlyMiddleware(t *testing.T) {
	t.Run(
		"finds the real Hijacker through a chain of Unwrap-only middleware writers", func(t *testing.T) {
			router := http.NewDefaultRouter(&errhttp.ErrorPipeline{}, http.ServeConfig{})
			router.Use(
				func(next nethttp.Handler) nethttp.Handler {
					return nethttp.HandlerFunc(
						func(w nethttp.ResponseWriter, r *nethttp.Request) {
							next.ServeHTTP(&unwrapOnlyResponseWriter{ResponseWriter: w}, r)
						},
					)
				},
			)
			router.Method(
				"GET", "/ws", nethttp.HandlerFunc(
					func(w nethttp.ResponseWriter, r *nethttp.Request) {
						w.Header().Set("Upgrade", "websocket")
						w.Header().Set("Connection", "Upgrade")
						w.WriteHeader(nethttp.StatusSwitchingProtocols)

						hj, ok := w.(nethttp.Hijacker)
						require.True(
							t, ok,
							"the buffer must see through the Unwrap-only wrapper to find a Hijacker",
						)

						conn, rw, err := hj.Hijack()
						require.NoError(t, err)
						defer conn.Close()

						_, err = rw.Write([]byte("raw-post-hijack-payload"))
						require.NoError(t, err)
						require.NoError(t, rw.Flush())
					},
				),
			)

			server := httptest.NewServer(router)
			defer server.Close()

			addr := server.Listener.Addr().String()
			response := rawRequest(
				t, addr,
				"GET /ws HTTP/1.1\r\nHost: "+addr+"\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n",
			)

			require.Contains(t, response, "101 Switching Protocols")
			require.Contains(t, response, "Upgrade: websocket")
			require.Contains(t, response, "raw-post-hijack-payload")
		},
	)
}

func TestDefaultRouter_RegularRequests(t *testing.T) {
	t.Run(
		"still serves normal responses through the buffer", func(t *testing.T) {
			router := http.NewDefaultRouter(&errhttp.ErrorPipeline{}, http.ServeConfig{})
			router.Method(
				"GET", "/ok", nethttp.HandlerFunc(
					func(w nethttp.ResponseWriter, r *nethttp.Request) {
						w.Header().Set("X-Test", "yes")
						w.WriteHeader(nethttp.StatusTeapot)
						_, _ = w.Write([]byte("teapot"))
					},
				),
			)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/ok", nil)
			router.ServeHTTP(rec, req)

			require.Equal(t, nethttp.StatusTeapot, rec.Code)
			require.Equal(t, "yes", rec.Header().Get("X-Test"))
			require.Equal(t, "teapot", rec.Body.String())
		},
	)

	t.Run(
		"still routes unmatched paths through the custom not-found handler", func(t *testing.T) {
			router := http.NewDefaultRouter(&errhttp.ErrorPipeline{}, http.ServeConfig{})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/does-not-exist", nil)
			router.ServeHTTP(rec, req)

			require.Equal(t, nethttp.StatusNotFound, rec.Code)
		},
	)
}
