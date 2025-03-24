package storage_test

import (
	"context"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDefaultIdentityRepository_Create(t *testing.T) {
	t.Parallel()
	t.Run(
		"data field is nil", func(t *testing.T) {
			t.Parallel()

			id := "test-identity-1"
			accountId := uuid.Must(uuid.NewV6())
			identity, err := identityRepository.Create(
				context.Background(),
				id,
				accountId,
				repository.IdentityTypeEmail,
				nil,
			)

			t.Log("Given the identity is not found")
			t.Log("When try to create")
			t.Log("	Then the identity is created")
			require.Nil(t, err)
			fixtureFactory.Identity().ID(identity.ID).Cleanup(t)
			require.Equal(t, id, identity.Identity)
			require.Equal(t, id, identity.Identity)
			require.Equal(t, accountId.String(), identity.AccountID.String())
			require.Equal(t, repository.IdentityTypeEmail, identity.Type)
			require.Equal(t, make(map[string]interface{}), identity.Data)
		},
	)

	t.Run(
		"data field is empty map", func(t *testing.T) {
			t.Parallel()

			id := "test-identity-2"
			accountId := uuid.Must(uuid.NewV6())
			identity, err := identityRepository.Create(
				context.Background(),
				id,
				accountId,
				repository.IdentityTypeEmail,
				map[string]interface{}{},
			)

			t.Log("Given the identity is not found")
			t.Log("When try to create")
			t.Log("	Then the identity is created")
			require.Nil(t, err)
			fixtureFactory.Identity().ID(identity.ID).Cleanup(t)
			require.Equal(t, id, identity.Identity)
			require.Equal(t, id, identity.Identity)
			require.Equal(t, accountId.String(), identity.AccountID.String())
			require.Equal(t, repository.IdentityTypeEmail, identity.Type)
			require.Equal(t, make(map[string]interface{}), identity.Data)
		},
	)

	t.Run(
		"data field is filled", func(t *testing.T) {
			t.Parallel()

			id := "test-identity-3"
			accountId := uuid.Must(uuid.NewV6())
			identity, err := identityRepository.Create(
				context.Background(),
				id,
				accountId,
				repository.IdentityTypeEmail,
				map[string]interface{}{
					"key1": "value1",
				},
			)

			t.Log("Given the identity is not found")
			t.Log("When try to create")
			t.Log("	Then the identity is created")
			require.Nil(t, err)
			fixtureFactory.Identity().ID(identity.ID).Cleanup(t)
			require.Equal(t, id, identity.Identity)
			require.Equal(t, id, identity.Identity)
			require.Equal(t, accountId.String(), identity.AccountID.String())
			require.Equal(t, repository.IdentityTypeEmail, identity.Type)
			require.Equal(t, "value1", identity.Data["key1"])
		},
	)
}
