package context

import (
	"context"
	"io"
	"net/http"
)

type ctxKeyRequest string

const RequestKey ctxKeyRequest = "request"

func GetRequest(ctx context.Context) *http.Request {
	if ctx == nil {
		return nil
	}
	if value := ctx.Value(RequestKey); value != nil {
		req, ok := value.(*http.Request)
		if !ok {
			return nil
		}
		if r, ok := req.Body.(io.Seeker); ok {
			_, _ = r.Seek(0, io.SeekStart)
		}
		return req
	}
	return nil
}

func WithRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, RequestKey, r)
}
