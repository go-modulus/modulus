package types

import (
	"context"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/gofrs/uuid"
)

func MarshalUuid(id uuid.UUID) graphql.ContextMarshaler {
	return graphql.ContextWriterFunc(
		func(_ context.Context, w io.Writer) error {
			_, _ = fmt.Fprintf(w, "%q", id.String())
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

	return uuid.Nil, erruser.NewValidationError(erruser.New(graphql.GetPath(ctx).String(), "Invalid UUID"))
}
