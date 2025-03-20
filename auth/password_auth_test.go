package auth_test

import (
	"context"
	"encoding/json"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPasswordAuthenticator_Register(t *testing.T) {
	t.Parallel()
	t.Run(
		"register identity without additional data", func(t *testing.T) {
			t.Parallel()
			account, err := passwordAuth.Register(
				context.Background(),
				"user",
				"password",
				repository.IdentityTypeNickname,
				[]string{},
				nil,
			)
			require.NoError(t, err)

			savedAccount := fixtureFactory.Account().ID(account.ID).PullUpdates(t).Cleanup(t).GetEntity()
			savedIdentity := fixtureFactory.Identity().AccountID(account.ID).PullUpdatesLastAccountIdentity(t).CleanupAllOfAccount(t).GetEntity()
			fixtureFactory.Credential().AccountID(account.ID).CleanupAllOfAccount(t)

			t.Log("When the account is registered")
			t.Log("	Then the account is returned")
			require.NoError(t, err)
			require.Equal(t, repository.AccountStatusActive, account.Status)

			t.Log("	And the account is saved")
			require.Equal(t, "user", savedIdentity.Identity)
			require.Equal(t, storage.IdentityStatusActive, savedIdentity.Status)

			t.Log("	And the account is created")
			require.Equal(t, account.ID, savedAccount.ID)
			require.Equal(t, storage.AccountStatusActive, savedAccount.Status)
		},
	)

	t.Run(
		"register identity with additional data", func(t *testing.T) {
			t.Parallel()
			account, err := passwordAuth.Register(
				context.Background(),
				"user1",
				"password",
				repository.IdentityTypeNickname,
				[]string{},
				map[string]interface{}{
					"key": "value",
				},
			)

			savedAccount := fixtureFactory.Account().ID(account.ID).PullUpdates(t).Cleanup(t).GetEntity()
			savedIdentity := fixtureFactory.Identity().AccountID(account.ID).PullUpdatesLastAccountIdentity(t).CleanupAllOfAccount(t).GetEntity()
			fixtureFactory.Credential().AccountID(account.ID).CleanupAllOfAccount(t)
			fixtureFactory.Identity().AccountID(account.ID).CleanupAllOfAccount(t)

			var data map[string]interface{}
			errUnmarshal := json.Unmarshal(savedIdentity.Data, &data)

			t.Log("When the account is registered")
			t.Log("	Then the account is returned")
			require.NoError(t, err)
			require.Equal(t, repository.AccountStatusActive, account.Status)

			t.Log("	And the account is saved")
			require.NoError(t, errUnmarshal)
			require.Equal(t, "user1", savedIdentity.Identity)
			require.Equal(t, storage.IdentityStatusActive, savedIdentity.Status)
			require.Equal(t, "value", data["key"])

			t.Log("	And the account is created")
			require.Equal(t, account.ID, savedAccount.ID)
			require.Equal(t, storage.AccountStatusActive, savedAccount.Status)
		},
	)

	t.Run(
		"fail on the second registration of identity", func(t *testing.T) {
			t.Parallel()
			account := fixtureFactory.Account().Create(t).GetEntity()
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				AccountID(account.ID).
				Identity("user2").
				Create(t).
				GetEntity()
			_, err := passwordAuth.Register(
				context.Background(),
				identity.Identity,
				"password",
				repository.IdentityTypeNickname,
				[]string{},
				nil,
			)

			t.Log("Given the identity is registered")
			t.Log("When the identity is registering again")
			t.Log("	Then the error ErrIdentityExists is returned")
			require.ErrorIs(t, err, repository.ErrIdentityExists)
		},
	)

	t.Run(
		"fail if identity is blocked", func(t *testing.T) {
			t.Parallel()
			accountId := uuid.Must(uuid.NewV6())
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				AccountID(accountId).
				Identity("user3").
				Status(storage.IdentityStatusBlocked).
				Create(t).
				GetEntity()

			_, err := passwordAuth.Register(
				context.Background(),
				identity.Identity,
				"password",
				repository.IdentityTypeNickname,
				[]string{},
				nil,
			)

			t.Log("Given the identity is blocked")
			t.Log("When the identity is registering again")
			t.Log("	Then the error ErrIdentityIsBlocked is returned")
			require.ErrorIs(t, err, auth.ErrIdentityIsBlocked)
		},
	)
}
func TestPasswordAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()
	t.Run(
		"authenticate", func(t *testing.T) {
			t.Parallel()
			identity := "user4"
			account, err := passwordAuth.Register(
				context.Background(),
				identity,
				"password",
				repository.IdentityTypeNickname,
				[]string{},
				map[string]interface{}{
					"key": "value",
				},
			)
			require.NoError(t, err)

			fixtureFactory.Account().ID(account.ID).Cleanup(t)
			fixtureFactory.Identity().AccountID(account.ID).CleanupAllOfAccount(t)
			fixtureFactory.Credential().AccountID(account.ID).CleanupAllOfAccount(t)

			performer, err := passwordAuth.Authenticate(
				context.Background(),
				identity,
				"password",
			)

			t.Log("Given the account is registered")
			t.Log("When try to authenticate with the correct password")
			t.Log("	Then the performer is returned")
			require.NoError(t, err)
			require.Equal(t, account.ID, performer.ID)
		},
	)

	t.Run(
		"fail if identity not exist", func(t *testing.T) {
			t.Parallel()

			_, err := passwordAuth.Authenticate(context.Background(), "user5", "password")

			t.Log("Given the identity is not registered")
			t.Log("When try to authenticate")
			t.Log("	Then the error ErrInvalidIdentity is returned")
			require.ErrorIs(t, err, auth.ErrInvalidIdentity)
		},
	)

	t.Run(
		"fail if no password in database", func(t *testing.T) {
			t.Parallel()

			account := fixtureFactory.Account().Create(t).GetEntity()
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				Identity("user6").
				AccountID(account.ID).
				Create(t).
				GetEntity()
			fixtureFactory.Credential().
				AccountID(account.ID).
				Type(string(repository.CredentialTypeOTP)).
				Hash("ssss").
				Create(t)
			_, err := passwordAuth.Authenticate(context.Background(), identity.Identity, "password")

			t.Log("Given the identity is registered")
			t.Log("Given credentials with the password type are not found")
			t.Log("When try to authenticate")
			t.Log("	Then the error ErrCredentialNotFound is returned")
			require.ErrorIs(t, err, auth.ErrInvalidPassword)
		},
	)

	t.Run(
		"fail if password is wrong", func(t *testing.T) {
			t.Parallel()

			userId := uuid.Must(uuid.NewV6())
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				Identity("user7").
				AccountID(userId).
				Create(t).
				GetEntity()

			fixtureFactory.Credential().
				AccountID(userId).
				Hash("ssss2").
				Create(t)
			_, err := passwordAuth.Authenticate(context.Background(), identity.Identity, "password")

			t.Log("Given the identity is registered")
			t.Log("Given the password is wrong")
			t.Log("When try to authenticate")
			t.Log("	Then the error ErrInvalidPassword is returned")
			require.ErrorIs(t, err, auth.ErrInvalidPassword)
		},
	)

	t.Run(
		"fail if identity is blocked", func(t *testing.T) {
			t.Parallel()

			userId := uuid.Must(uuid.NewV6())
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				Identity("user8").
				AccountID(userId).
				Status(storage.IdentityStatusBlocked).
				Create(t).
				GetEntity()

			_, err := passwordAuth.Authenticate(context.Background(), identity.Identity, "password")

			t.Log("Given the identity is blocked")
			t.Log("When try to authenticate")
			t.Log("	Then the error ErrIdentityIsBlocked is returned")
			require.ErrorIs(t, err, auth.ErrIdentityIsBlocked)
		},
	)
}

func TestPasswordAuthenticator_RemoveIdentity(t *testing.T) {
	t.Parallel()
	t.Run(
		"remove identity with account", func(t *testing.T) {
			t.Parallel()
			account := fixtureFactory.Account().Create(t).GetEntity()
			identity := fixtureFactory.Identity().AccountID(account.ID).Create(t).GetEntity()

			err := passwordAuth.RemoveIdentity(context.Background(), identity.Identity)

			_, errIdent := identityRepository.Get(context.Background(), identity.Identity)
			_, errAcc := accountRepository.Get(context.Background(), account.ID)

			t.Log("Given only one identity is registered")
			t.Log("  When the identity is removed")
			t.Log("  Then both identity and account are removed")
			require.NoError(t, err)
			require.ErrorIs(t, errIdent, repository.ErrIdentityNotFound)
			require.ErrorIs(t, errAcc, repository.ErrAccountNotFound)
		},
	)

	t.Run(
		"remove identity only", func(t *testing.T) {
			t.Parallel()
			account := fixtureFactory.Account().Create(t).GetEntity()
			identity := fixtureFactory.Identity().AccountID(account.ID).Create(t).GetEntity()
			identity2 := fixtureFactory.Identity().AccountID(account.ID).Create(t).GetEntity()

			err := passwordAuth.RemoveIdentity(context.Background(), identity.Identity)

			_, errIdent := identityRepository.Get(context.Background(), identity.Identity)
			_, errIdent2 := identityRepository.Get(context.Background(), identity2.Identity)
			_, errAcc := accountRepository.Get(context.Background(), account.ID)

			t.Log("Given more than one identity is registered")
			t.Log("  When the identity is removed")
			t.Log("  Then both identity and account are removed")
			require.NoError(t, err)
			require.ErrorIs(t, errIdent, repository.ErrIdentityNotFound)
			require.NoError(t, errAcc)
			require.NoError(t, errIdent2)
		},
	)
}
