package types

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"io"
)

const Void = iota

func MarshalVoid(_ interface{}) graphql.ContextMarshaler {
	return graphql.ContextWriterFunc(func(_ context.Context, w io.Writer) error {
		_, _ = w.Write([]byte("null"))
		return nil
	})
}

func UnmarshalVoid(ctx context.Context, value interface{}) (any, error) {
	return nil, nil
}
