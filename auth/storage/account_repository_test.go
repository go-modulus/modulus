package storage_test

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDefaultAccountRepository_Create(t *testing.T) {
	t.Parallel()
	t.Run(
		"It should create a new account without roles and data", func(t *testing.T) {
			t.Parallel()

			id := uuid.Must(uuid.NewV6())
			account, err := accountRepository.Create(
				context.Background(),
				id,
				nil,
				nil,
			)

			t.Log("Given the account is not found")
			t.Log("When try to create")
			t.Log("	Then the account is created")
			require.Nil(t, err)
			fixtureFactory.Account().ID(account.ID).Cleanup(t)
			require.Equal(t, id.String(), account.ID.String())
			require.Equal(t, []string{}, account.Roles)
			require.Equal(t, make(map[string]interface{}), account.Data)
		},
	)

	t.Run(
		"It should create a new account with empty roles and data", func(t *testing.T) {
			t.Parallel()

			id := uuid.Must(uuid.NewV6())
			account, err := accountRepository.Create(
				context.Background(),
				id,
				[]string{},
				map[string]interface{}{},
			)

			t.Log("Given the account is not found")
			t.Log("When try to create with empty roles and data")
			t.Log("	Then the account is created")
			require.Nil(t, err)
			fixtureFactory.Account().ID(account.ID).Cleanup(t)
			require.Equal(t, id.String(), account.ID.String())
			require.Equal(t, []string{}, account.Roles)
			require.Equal(t, make(map[string]interface{}), account.Data)
		},
	)

	t.Run(
		"It should create a new account with roles and data", func(t *testing.T) {
			t.Parallel()

			id := uuid.Must(uuid.NewV6())
			account, err := accountRepository.Create(
				context.Background(),
				id,
				[]string{"role1", "role2"},
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			)

			t.Log("Given the account is not found")
			t.Log("When try to create with roles and data")
			t.Log("	Then the account is created")
			require.Nil(t, err)
			fixtureFactory.Account().ID(account.ID).Cleanup(t)
			require.Equal(t, id.String(), account.ID.String())
			require.Equal(t, []string{"role1", "role2"}, account.Roles)
			require.Equal(t, "value1", account.Data["key1"])
			require.Equal(t, "value2", account.Data["key2"])
		},
	)
}
