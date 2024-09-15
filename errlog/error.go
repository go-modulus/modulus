package errlog

import (
	"braces.dev/errtrace"
	slogformatter "github.com/samber/slog-formatter"
	_ "golang.org/x/text/message"
	"log/slog"
	"reflect"
	"strings"
)

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

func Formatter() slogformatter.Formatter {
	return slogformatter.FormatByType[error](
		func(err error) slog.Value {
			values := getErrorAttrs(err)
			cause := Cause(err)
			if cause != nil {
				causeValues := getErrorAttrs(cause)
				values = append(values, slog.Any("cause", causeValues))
			}

			return slog.GroupValue(values...)
		},
	)
}

func getErrorAttrs(err error) []slog.Attr {
	message := errtrace.FormatString(err)
	msgParts := strings.Split(message, "\n")
	message = msgParts[0]
	var trace []string
	for i, part := range msgParts {
		if i == 0 || part == "" {
			continue
		}
		trace = append(trace, part)
	}

	meta := Meta(err)
	metaValues := make([]any, 0, len(meta))
	for k, v := range meta {
		metaValues = append(metaValues, slog.String(k, v))
	}
	values := []slog.Attr{
		slog.String("type", reflect.TypeOf(err).String()),
		slog.String("message", message),
		slog.Any("trace", trace),
		slog.Group("meta", metaValues...),
	}
	return values
}
