package graphql_test

import (
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/install/graphql"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuthGuard(t *testing.T) {
	t.Parallel()
	t.Run(
		"pass authenticated user", func(t *testing.T) {
			t.Parallel()

			ctx := auth.WithPerformer(
				context.Background(),
				auth.Performer{
					ID:        uuid.Must(uuid.NewV6()),
					SessionID: uuid.Must(uuid.NewV4()),
				},
			)

			res, err := graphql.AuthGuard(
				ctx,
				nil,
				func(ctx context.Context) (res interface{}, err error) {
					return "ok", nil
				},
				[]string{},
			)
			require.NoError(t, err)
			require.Equal(t, "ok", res)
		},
	)

	t.Run(
		"pass authenticated user with roles", func(t *testing.T) {
			t.Parallel()

			ctx := auth.WithPerformer(
				context.Background(),
				auth.Performer{
					ID:        uuid.Must(uuid.NewV6()),
					SessionID: uuid.Must(uuid.NewV4()),
					Roles:     []string{"admin"},
				},
			)

			res, err := graphql.AuthGuard(
				ctx,
				nil,
				func(ctx context.Context) (res interface{}, err error) {
					return "ok", nil
				},
				[]string{"admin"},
			)
			require.NoError(t, err)
			require.Equal(t, "ok", res)
		},
	)

	t.Run(
		"fail authenticated user without allowed role", func(t *testing.T) {
			t.Parallel()

			ctx := auth.WithPerformer(
				context.Background(),
				auth.Performer{
					ID:        uuid.Must(uuid.NewV6()),
					SessionID: uuid.Must(uuid.NewV4()),
					Roles:     []string{"user"},
				},
			)

			_, err := graphql.AuthGuard(
				ctx,
				nil,
				func(ctx context.Context) (res interface{}, err error) {
					return "ok", nil
				},
				[]string{"admin"},
			)
			require.ErrorIs(t, err, auth.ErrUnauthorized)
		},
	)

	t.Run(
		"fail unauthenticated user", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			_, err := graphql.AuthGuard(
				ctx,
				nil,
				func(ctx context.Context) (res interface{}, err error) {
					return "ok", nil
				},
				[]string{},
			)
			require.ErrorIs(t, err, auth.ErrUnauthenticated)
		},
	)
}
