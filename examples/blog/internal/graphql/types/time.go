package types

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/go-modulus/modulus/errors/erruser"
	"io"
	"time"
)

func MarshalTime(value time.Time) graphql.ContextMarshaler {
	return graphql.ContextWriterFunc(
		func(_ context.Context, w io.Writer) error {
			_, _ = w.Write([]byte(fmt.Sprintf("%q", value.Format(time.RFC3339))))
			return nil
		},
	)
}

func UnmarshalTime(ctx context.Context, value interface{}) (time.Time, error) {
	rawTime, ok := value.(string)
	if ok {
		t, err := time.Parse(time.RFC3339, rawTime)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, erruser.NewValidationError(
		erruser.New(
			graphql.GetPath(ctx).String(),
			"Invalid Time. Pass it in format "+time.RFC3339,
		),
	)
}
