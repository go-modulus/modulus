package types

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/go-modulus/modulus/validator"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
	"io"
)

func MarshalUuid(id uuid.UUID) graphql.ContextMarshaler {
	return graphql.ContextWriterFunc(
		func(_ context.Context, w io.Writer) error {
			_, _ = w.Write([]byte(fmt.Sprintf("%q", id.String())))
			return nil
		},
	)
}

func UnmarshalUuid(ctx context.Context, value interface{}) (uuid.UUID, error) {
	rawUuid, ok := value.(string)
	if ok {
		id, err := uuid.FromString(rawUuid)
		if err == nil {
			return id, nil
		}
	}

	return uuid.Nil, &validator.ErrInvalidInput{
		Fields: []validator.InvalidField{
			validator.NewInvalidFieldFromOzzo(validator.Path(ctx), is.ErrUUID),
		},
	}
}
