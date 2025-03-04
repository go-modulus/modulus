package errhttp

import (
	"encoding/json"
	"fmt"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errlog"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/http/context"
	"log/slog"
	"net/http"
)

func SendError(
	logger *slog.Logger,
	w http.ResponseWriter,
	req *http.Request,
	err error,
) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	code := erruser.Code(err)
	status := Status(err)
	if code == erruser.InternalErrorCode {
		logger.ErrorContext(req.Context(), "unhandled error", errlog.Error(err))
	} else if status == http.StatusInternalServerError {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)

	message := errors.Hint(err)
	details := erruser.Details(err)

	requestID := context.GetRequestID(req.Context())
	if requestID != "" {
		if details == nil {
			details = make(map[string]any)
		}
		details["requestId"] = requestID
		if code == erruser.InternalErrorCode {
			message = fmt.Sprintf("%s (RID: %s)", message, requestID)
		}
	}

	_ = json.NewEncoder(w).Encode(
		map[string]interface{}{
			"error": map[string]any{
				"code":    code,
				"message": message,
				"details": details,
			},
		},
	)
}

func WrapHandler(logger *slog.Logger, handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				if p == http.ErrAbortHandler {
					panic(p)
				}

				err, ok := p.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", p)
				}
				SendError(logger, w, req, err)
			}
		}()

		err := handler(w, req)
		if err != nil {
			SendError(logger, w, req, err)
		}
	}
}

func WrapMiddleware(
	logger *slog.Logger,
	middleware func(http.Handler) Handler,
) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return WrapHandler(logger, middleware(handler))
	}
}
