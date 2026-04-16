package errlog

import (
	"log/slog"
	"reflect"

	"github.com/go-modulus/modulus/errors"
	_ "golang.org/x/text/message"
)

type errorValue struct {
	err error
}

func (e errorValue) LogValue() slog.Value {
	values := getErrorAttrs(e.err)
	return slog.GroupValue(values...)
}

func Error(err error) slog.Attr {
	return slog.Any("error", errorValue{err: err})
}

func getErrorAttrs(err error) []slog.Attr {
	message := err.Error()
	meta := errors.Meta(err)
	trace := errors.Trace(err)
	hint := errors.Hint(err)
	metaValues := make([]any, 0, len(meta))
	for k, v := range meta {
		metaValues = append(metaValues, slog.String(k, v))
	}
	values := []slog.Attr{
		slog.String("type", reflect.TypeOf(err).String()),
		slog.String("message", message),
		slog.String("hint", hint),
		slog.Any("trace", trace),
		slog.Group("meta", metaValues...),
	}
	cause := errors.Cause(err)
	if cause != nil {
		if errors.InternalErrorCode != cause.Error() {
			values = append(values, slog.Any("cause", getErrorAttrs(cause)))
		} else {
			cause = errors.Cause(cause)
			if cause != nil {
				values = append(values, slog.Any("cause", getErrorAttrs(cause)))
			}
		}
	}
	return values
}
