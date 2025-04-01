package errhttp

import (
	"encoding/json"
	"fmt"
	"github.com/go-modulus/modulus/errors"
	"net/http"
	"strconv"
)

const HttpCodeMetaName = "httpCode"

func ErrWithHttpCode(
	err error,
	code int,
) error {
	return errors.WithAddedMeta(err, HttpCodeMetaName, fmt.Sprintf("%d", code))
}

func HttpCode(
	err error,
) int {
	status := http.StatusInternalServerError
	if errors.IsUserError(err) {
		status = http.StatusBadRequest
	}
	meta := errors.Meta(err)
	if httpCode, ok := meta[HttpCodeMetaName]; ok {
		tmpStatus, err := strconv.Atoi(httpCode)
		delete(meta, HttpCodeMetaName)
		if err == nil {
			status = tmpStatus
		}
	}

	return status
}

func SendError(
	w http.ResponseWriter,
	err error,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	code := err.Error()
	message := errors.Hint(err)
	meta := errors.Meta(err)
	httpCode := HttpCode(err)

	w.WriteHeader(httpCode)

	_ = json.NewEncoder(w).Encode(
		map[string]interface{}{
			"error": map[string]any{
				"message": message,
				"extensions": map[string]any{
					"code": code,
					"meta": meta,
				},
			},
		},
	)
}

func WrapHandler(errorPipeline *ErrorPipeline, handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		defer func() {
			if p := recover(); p != nil {
				if p == http.ErrAbortHandler {
					panic(p)
				}

				err, ok := p.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", p)
				}

				err = errorPipeline.Process(ctx, err)
				SendError(w, err)
			}
		}()

		err := handler(w, req)
		if err != nil {
			err = errorPipeline.Process(ctx, err)
			SendError(w, err)
		}
	}
}

func WrapMiddleware(
	errorPipeline *ErrorPipeline,
	middleware func(http.Handler) Handler,
) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return WrapHandler(errorPipeline, middleware(handler))
	}
}
