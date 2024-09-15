package logger

import (
	"context"
	"log/slog"
)

type key int

var tagsKey key

func AddTags(ctx context.Context, kv ...string) context.Context {
	return context.WithValue(ctx, tagsKey, createTags(ctx, kv...))
}

func TagsFromContext(ctx context.Context) (map[string]string, bool) {
	meta, ok := ctx.Value(tagsKey).(map[string]string)
	return meta, ok
}

func createTags(ctx context.Context, kv ...string) map[string]string {
	meta := make(map[string]string)

	if parent, ok := TagsFromContext(ctx); ok {
		// make a copy to avoid mutating parent context meta via map reference.
		for k, v := range parent {
			meta[k] = v
		}
	}

	l := len(kv)
	if l%2 != 0 {
		l -= 1 // don't error on odd number of args
	}

	for i := 0; i < l; i += 2 {
		k := kv[i]
		v := kv[i+1]

		meta[k] = v
	}

	return meta
}

func Tags(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	tags, ok := TagsFromContext(ctx)
	if ok {
		attrs := make([]slog.Attr, 0, len(tags))
		for k, v := range tags {
			attrs = append(attrs, slog.String(k, v))
		}
		if len(attrs) > 0 {
			record.AddAttrs(attrs...)
		}
	}
	return next(ctx, record)
}
