package errlog

import (
	"log/slog"
	"reflect"

	"github.com/go-modulus/modulus/errors"
	slogformatter "github.com/samber/slog-formatter"
	_ "golang.org/x/text/message"
)

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

func Formatter() slogformatter.Formatter {
	return slogformatter.FormatByType[error](
		func(err error) slog.Value {
			values := getErrorAttrs(err)
			cause := errors.Cause(err)
			if cause != nil {
				causeValues := getErrorAttrs(cause)
				values = append(values, slog.Any("cause", causeValues))
			}

			return slog.GroupValue(values...)
		},
	)
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
	return values
}
