package types

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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

	return time.Time{}, &validator.ErrInvalidInput{
		Fields: []validator.InvalidField{
			validator.NewInvalidFieldFromOzzo(validator.Path(ctx), validation.ErrDateInvalid),
		},
	}
}
