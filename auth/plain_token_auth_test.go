package auth_test

import (
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
	"testing"
	"time"
)

func TestPlainTokenAuthenticator_StartSession(t *testing.T) {
	t.Parallel()
	t.Run(
		"should return a valid pair", func(t *testing.T) {
			t.Parallel()
			identity := fixtureFactory.Identity().Create(t).GetEntity()

			pair, err := plainTokenAuth.IssueTokens(
				context.Background(),
				identity.ID,
				nil,
			)
			at := pair.AccessToken
			rt := pair.RefreshToken

			savedAt := fixtureFactory.AccessToken().Hash(at.Hash).PullUpdates(t).Cleanup(t).GetEntity()
			savedRt := fixtureFactory.RefreshToken().Hash(rt.Hash).PullUpdates(t).Cleanup(t).GetEntity()

			t.Log("When the session is started")
			t.Log(" Then the access and refresh tokens should be created")
			require.NoError(t, err)
			require.Equal(t, at.UserID, savedAt.UserID)
			require.Equal(t, at.SessionID, savedAt.SessionID)
			require.Equal(t, rt.SessionID, savedRt.SessionID)

			t.Log(" And the access token should expire in 1 hour")
			require.WithinDuration(t, time.Now().Add(time.Hour), savedAt.ExpiresAt, 2*time.Second)

			t.Log(" And the refresh token should expire in 720 hours")
			require.WithinDuration(t, time.Now().Add(720*time.Hour), savedRt.ExpiresAt, 2*time.Second)
		},
	)
}

func TestPlainTokenAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()
	t.Run(
		"should return a valid performer", func(t *testing.T) {
			t.Parallel()
			identity := fixtureFactory.Identity().Create(t).GetEntity()

			pair, err := plainTokenAuth.IssueTokens(
				context.Background(),
				identity.ID,
				nil,
			)

			require.NoError(t, err)

			performer, err := plainTokenAuth.Authenticate(
				context.Background(),
				pair.AccessToken.Token.String,
			)
			at := pair.AccessToken
			rt := pair.RefreshToken

			fixtureFactory.AccessToken().Hash(at.Hash).Cleanup(t)
			fixtureFactory.RefreshToken().Hash(rt.Hash).Cleanup(t)

			t.Log("Given the valid access token")
			t.Log("When authenticate the user")
			t.Log(" Then valid performer should be returned")
			require.NoError(t, err)
			require.Equal(t, at.UserID, performer.ID)
			require.Equal(t, at.SessionID, performer.SessionID)
			require.Equal(t, identity.Roles, performer.Roles)
		},
	)

	t.Run(
		"should return an error if the token is revoked", func(t *testing.T) {
			t.Parallel()
			token := "test1"
			hash := hashStrategy.Hash(token)
			fixtureFactory.AccessToken().
				RevokedAt(null.TimeFrom(time.Now())).
				Hash(hash).
				Create(t)

			_, err := plainTokenAuth.Authenticate(
				context.Background(),
				token,
			)

			t.Log("Given the revoked access token")
			t.Log("When authenticate the user")
			t.Log(" Then an error should be returned")
			require.ErrorIs(t, err, auth.ErrTokenIsRevoked)
		},
	)

	t.Run(
		"should return an error if the token is expired", func(t *testing.T) {
			t.Parallel()
			token := "test2"
			hash := hashStrategy.Hash(token)
			fixtureFactory.AccessToken().
				ExpiresAt(time.Now().Add(-1 * time.Hour)).
				Hash(hash).
				Create(t)

			_, err := plainTokenAuth.Authenticate(
				context.Background(),
				token,
			)

			t.Log("Given the expired access token")
			t.Log("When authenticate the user")
			t.Log(" Then an error should be returned")
			require.ErrorIs(t, err, auth.ErrTokenIsExpired)
		},
	)

	t.Run(
		"should return an error if the token is not exist", func(t *testing.T) {
			t.Parallel()
			_, err := plainTokenAuth.Authenticate(
				context.Background(),
				"test3",
			)

			t.Log("Given the expired access token")
			t.Log("When authenticate the user")
			t.Log(" Then an error should be returned")
			require.ErrorIs(t, err, repository.ErrTokenNotExist)
		},
	)
}

func TestPlainTokenAuthenticator_IssueNewAccessToken(t *testing.T) {
	t.Parallel()
	t.Run(
		"should return a new access token", func(t *testing.T) {
			t.Parallel()
			identity := fixtureFactory.Identity().Create(t).GetEntity()

			pair, err := plainTokenAuth.IssueTokens(
				context.Background(),
				identity.ID,
				nil,
			)

			require.NoError(t, err)

			at, err := plainTokenAuth.IssueNewAccessToken(
				context.Background(),
				pair.RefreshToken.Token.String,
				nil,
			)
			require.NoError(t, err)

			performer, err := plainTokenAuth.Authenticate(
				context.Background(),
				at.Token.String,
			)

			rt := pair.RefreshToken

			fixtureFactory.AccessToken().Hash(at.Hash).Cleanup(t)
			fixtureFactory.AccessToken().Hash(pair.AccessToken.Hash).Cleanup(t)
			fixtureFactory.RefreshToken().Hash(rt.Hash).Cleanup(t)

			t.Log("Given the valid new access token")
			t.Log("When authenticate the user")
			t.Log(" Then valid performer should be returned")
			require.NoError(t, err)
			require.Equal(t, at.UserID, performer.ID)
			require.Equal(t, at.SessionID, performer.SessionID)
			require.Equal(t, identity.Roles, performer.Roles)
		},
	)
}

func TestPlainTokenAuthenticator_RefreshAccessToken(t *testing.T) {
	t.Parallel()
	t.Run(
		"should return a new access token", func(t *testing.T) {
			t.Parallel()
			identity := fixtureFactory.Identity().Create(t).GetEntity()

			pair, err := plainTokenAuth.IssueTokens(
				context.Background(),
				identity.ID,
				nil,
			)

			require.NoError(t, err)

			at, err := plainTokenAuth.RefreshAccessToken(
				context.Background(),
				pair.RefreshToken.Token.String,
				nil,
				-1*time.Second,
			)
			require.NoError(t, err)

			performer, err := plainTokenAuth.Authenticate(
				context.Background(),
				at.Token.String,
			)

			rt := pair.RefreshToken

			fixtureFactory.AccessToken().Hash(at.Hash).Cleanup(t)
			oldToken := fixtureFactory.AccessToken().Hash(pair.AccessToken.Hash).Cleanup(t).PullUpdates(t).GetEntity()
			fixtureFactory.RefreshToken().Hash(rt.Hash).Cleanup(t)

			t.Log("Given the valid refreshed access token")
			t.Log("When authenticate the user")
			t.Log(" Then valid performer should be returned")
			require.NoError(t, err)
			require.Equal(t, at.UserID, performer.ID)
			require.Equal(t, at.SessionID, performer.SessionID)
			require.Equal(t, identity.Roles, performer.Roles)
			require.True(t, oldToken.ExpiresAt.Before(time.Now()))
		},
	)
}
