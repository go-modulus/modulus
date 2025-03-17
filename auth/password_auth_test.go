package auth_test

import (
	"context"
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
			userId := uuid.Must(uuid.NewV6())
			identity, err := passwordAuth.Register(
				context.Background(),
				"user",
				"password",
				userId,
				[]string{},
				nil,
			)
			require.NoError(t, err)

			savedIdentity := fixtureFactory.Identity().ID(identity.ID).PullUpdates(t).Cleanup(t).GetEntity()
			fixtureFactory.Credential().IdentityID(identity.ID).CleanupAllOfIdentity(t)

			t.Log("When the identity is registered")
			t.Log("	Then the identity is returned")
			require.NoError(t, err)
			require.Equal(t, userId, identity.UserID)
			require.Equal(t, "user", identity.Identity)
			require.Equal(t, repository.IdentityStatusActive, identity.Status)
			require.Empty(t, identity.Data)

			t.Log("	And the identity is saved")
			require.Equal(t, identity.UserID, savedIdentity.UserID)
			require.Equal(t, identity.Identity, savedIdentity.Identity)
			require.Equal(t, storage.IdentityStatusActive, savedIdentity.Status)
		},
	)

	t.Run(
		"register identity with additional data", func(t *testing.T) {
			t.Parallel()
			userId := uuid.Must(uuid.NewV6())
			identity, err := passwordAuth.Register(
				context.Background(),
				"user1",
				"password",
				userId,
				[]string{},
				map[string]interface{}{
					"key": "value",
				},
			)

			savedIdentity := fixtureFactory.Identity().ID(identity.ID).PullUpdates(t).GetEntity()
			fixtureFactory.Credential().IdentityID(identity.ID).CleanupAllOfIdentity(t)
			fixtureFactory.Identity().UserID(userId).CleanupAllOfUser(t)

			t.Log("When the identity is registered")
			t.Log("	Then the identity is returned")
			require.NoError(t, err)
			require.Equal(t, userId, identity.UserID)
			require.Equal(t, "user1", identity.Identity)
			require.Equal(t, repository.IdentityStatusActive, identity.Status)
			require.Equal(t, "value", identity.Data["key"])

			t.Log("	And the identity is saved")
			require.Equal(t, identity.UserID, savedIdentity.UserID)
			require.Equal(t, identity.Identity, savedIdentity.Identity)
			require.Equal(t, storage.IdentityStatusActive, savedIdentity.Status)
		},
	)

	t.Run(
		"fail on the second registration of identity", func(t *testing.T) {
			t.Parallel()
			userId := uuid.Must(uuid.NewV6())
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				UserID(userId).
				Identity("user2").
				Create(t).
				GetEntity()
			_, err := passwordAuth.Register(
				context.Background(),
				identity.Identity,
				"password",
				userId,
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
			userId := uuid.Must(uuid.NewV6())
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				UserID(userId).
				Identity("user3").
				Status(storage.IdentityStatusBlocked).
				Create(t).
				GetEntity()

			_, err := passwordAuth.Register(
				context.Background(),
				identity.Identity,
				"password",
				userId,
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
			userId := uuid.Must(uuid.NewV6())
			identity, err := passwordAuth.Register(
				context.Background(),
				"user4",
				"password",
				userId,
				[]string{},
				map[string]interface{}{
					"key": "value",
				},
			)
			require.NoError(t, err)

			fixtureFactory.Identity().ID(identity.ID).Cleanup(t)
			fixtureFactory.Credential().IdentityID(identity.ID).CleanupAllOfIdentity(t)

			performer, err := passwordAuth.Authenticate(
				context.Background(),
				identity.Identity,
				"password",
			)

			t.Log("Given the identity is registered")
			t.Log("When try to authenticate with the correct password")
			t.Log("	Then the performer is returned")
			require.NoError(t, err)
			require.Equal(t, userId, performer.ID)
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

			userId := uuid.Must(uuid.NewV6())
			identity := fixtureFactory.Identity().
				ID(uuid.Must(uuid.NewV6())).
				Identity("user6").
				UserID(userId).
				Create(t).
				GetEntity()
			fixtureFactory.Credential().
				IdentityID(identity.ID).
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
				UserID(userId).
				Create(t).
				GetEntity()

			fixtureFactory.Credential().
				IdentityID(identity.ID).
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
				UserID(userId).
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
