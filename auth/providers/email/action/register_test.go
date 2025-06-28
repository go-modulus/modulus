package action_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/go-modulus/modulus/auth/providers/email/action"
)

func TestRegisterUser_Execute(t *testing.T) {
	t.Run(
		"Success", func(t *testing.T) {
			ctx := context.Background()

			creatorMock.Test(t)
			t.Cleanup(
				func() {
					creatorMock.AssertExpectations(t)
					creatorMock.ExpectedCalls = make([]*mock.Call, 0)
					creatorMock.Calls = make([]mock.Call, 0)
				},
			)

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "@gmail.com",
				Password: gofakeit.Password(true, true, true, true, false, 20),
			}
			creatorMock.On(
				"CreateUser",
				mock.Anything,
				mock.MatchedBy(
					func(u action.User) bool {
						return u.Email == request.Email
					},
				),
			).Return(action.User{}, nil)

			pair, err := register.Execute(ctx, request)
			require.NoError(t, err, "Pair should be created")

			account := authFixture.Account().ID(pair.AccessToken.AccountID).PullUpdates(t).Cleanup(t).GetEntity()
			identity := authFixture.Identity().ID(pair.AccessToken.IdentityID).PullUpdates(t).Cleanup(t).GetEntity()
			accessToken := authFixture.AccessToken().Hash(pair.AccessToken.Hash).PullUpdates(t).Cleanup(t).GetEntity()
			refreshToken := authFixture.RefreshToken().Hash(pair.RefreshToken.Hash).PullUpdates(t).Cleanup(t).GetEntity()

			t.Log("Given the email is not registered as identity")
			t.Log("When the email is registering")
			t.Log("	Then the identity is registered successfully")
			require.Nil(t, err)
			require.NotEmpty(t, identity.ID)
			require.Equal(t, request.Email, identity.Identity)

			t.Log("   And account is created")
			require.Equal(t, account.ID.String(), identity.AccountID.String())
			require.Equal(t, storage.AccountStatusActive, account.Status)

			t.Log("   And account has default role")
			require.Len(t, account.Roles, 1)
			require.Equal(t, action.DefaultUserRole, account.Roles[0])

			t.Log("   And the tokens are created")
			require.Equal(t, pair.AccessToken.AccountID.String(), identity.AccountID.String())
			require.Equal(t, pair.AccessToken.SessionID.String(), accessToken.SessionID.String())
			require.Equal(t, pair.RefreshToken.IdentityID.String(), identity.ID.String())
			require.Equal(t, pair.RefreshToken.SessionID.String(), refreshToken.SessionID.String())

		},
	)

	t.Run(
		"register when email has plus", func(t *testing.T) {
			ctx := context.Background()

			creatorMock.Test(t)
			t.Cleanup(
				func() {
					creatorMock.AssertExpectations(t)
					creatorMock.ExpectedCalls = make([]*mock.Call, 0)
					creatorMock.Calls = make([]mock.Call, 0)
				},
			)

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "+test@gmail.com",
				Password: gofakeit.Password(true, true, true, true, false, 20),
				Roles:    []string{"test-role"},
			}
			creatorMock.On(
				"CreateUser",
				mock.Anything,
				mock.MatchedBy(
					func(u action.User) bool {
						return assert.Equal(t, u.Email, request.Email, "Email is not equal")
					},
				),
			).Return(action.User{}, nil)

			pair, err := register.Execute(ctx, request)
			require.NoError(t, err, "Pair should be created")

			account := authFixture.Account().ID(pair.AccessToken.AccountID).PullUpdates(t).Cleanup(t).GetEntity()
			identity := authFixture.Identity().ID(pair.AccessToken.IdentityID).PullUpdates(t).Cleanup(t).GetEntity()
			accessToken := authFixture.AccessToken().Hash(pair.AccessToken.Hash).PullUpdates(t).Cleanup(t).GetEntity()
			refreshToken := authFixture.RefreshToken().Hash(pair.RefreshToken.Hash).PullUpdates(t).Cleanup(t).GetEntity()

			t.Log("Given the email is not registered as identity")
			t.Log("When the email is registering")
			t.Log("	Then the identity is registered successfully")
			require.Nil(t, err)
			require.NotEmpty(t, identity.ID)
			require.Equal(t, request.Email, identity.Identity)

			t.Log("   And account is created")
			require.Equal(t, account.ID.String(), identity.AccountID.String())
			require.Equal(t, storage.AccountStatusActive, account.Status)
			require.Equal(t, request.Roles, account.Roles)

			t.Log("   And the tokens are created")
			require.Equal(t, pair.AccessToken.AccountID.String(), identity.AccountID.String())
			require.Equal(t, pair.AccessToken.SessionID.String(), accessToken.SessionID.String())
			require.Equal(t, pair.RefreshToken.IdentityID.String(), identity.ID.String())
			require.Equal(t, pair.RefreshToken.SessionID.String(), refreshToken.SessionID.String())

		},
	)
	t.Run(
		"register with user info", func(t *testing.T) {
			ctx := context.Background()

			creatorMock.Test(t)
			t.Cleanup(
				func() {
					creatorMock.AssertExpectations(t)
					creatorMock.ExpectedCalls = make([]*mock.Call, 0)
					creatorMock.Calls = make([]mock.Call, 0)
				},
			)

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "+test@gmail.com",
				Password: gofakeit.Password(true, true, true, true, false, 20),
				UserInfo: map[string]interface{}{
					"test": "test-val",
				},
				Roles: []string{"test-role"},
			}
			creatorMock.On(
				"CreateUser",
				mock.Anything,
				mock.MatchedBy(
					func(u action.User) bool {
						return assert.Equal(t, u.Email, request.Email, "Email is not equal") &&
							assert.Equal(t, u.UserInfo["test"], request.UserInfo["test"], "UserInfo is not equal")
					},
				),
			).Return(action.User{}, nil)

			pair, err := register.Execute(ctx, request)
			require.NoError(t, err, "Pair should be created")

			account := authFixture.Account().ID(pair.AccessToken.AccountID).PullUpdates(t).Cleanup(t).GetEntity()
			identity := authFixture.Identity().ID(pair.AccessToken.IdentityID).PullUpdates(t).Cleanup(t).GetEntity()
			accessToken := authFixture.AccessToken().Hash(pair.AccessToken.Hash).PullUpdates(t).Cleanup(t).GetEntity()
			refreshToken := authFixture.RefreshToken().Hash(pair.RefreshToken.Hash).PullUpdates(t).Cleanup(t).GetEntity()

			var userInfo map[string]interface{}
			errJson := json.Unmarshal(account.Data, &userInfo)

			t.Log("Given the email is not registered as identity")
			t.Log("When the email is registering")
			t.Log("	Then the identity is registered successfully")
			require.Nil(t, err)
			require.NotEmpty(t, identity.ID)
			require.Equal(t, request.Email, identity.Identity)
			require.NoError(t, errJson)
			require.Equal(t, request.UserInfo["test"], userInfo["test"])

			t.Log("   And account is created")
			require.Equal(t, account.ID.String(), identity.AccountID.String())
			require.Equal(t, request.Roles, account.Roles)

			t.Log("   And the tokens are created")
			require.Equal(t, pair.AccessToken.AccountID.String(), identity.AccountID.String())
			require.Equal(t, pair.AccessToken.SessionID.String(), accessToken.SessionID.String())
			require.Equal(t, pair.RefreshToken.IdentityID.String(), identity.ID.String())
			require.Equal(t, pair.RefreshToken.SessionID.String(), refreshToken.SessionID.String())

		},
	)

	t.Run(
		"email already exists in the external user table", func(t *testing.T) {
			ctx := context.Background()

			creatorMock.Test(t)
			t.Cleanup(
				func() {
					creatorMock.AssertExpectations(t)
					creatorMock.ExpectedCalls = make([]*mock.Call, 0)
					creatorMock.Calls = make([]mock.Call, 0)
				},
			)

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "@gmail.com",
				Password: gofakeit.Password(true, true, true, true, false, 20),
			}
			creatorMock.On(
				"CreateUser",
				mock.Anything,
				mock.MatchedBy(
					func(u action.User) bool {
						return assert.Equal(t, u.Email, request.Email, "Email is not equal")
					},
				),
			).Return(
				action.User{
					ID: uuid.Must(uuid.NewV4()),
				}, action.ErrUserAlreadyExists,
			)

			_, err := register.Execute(ctx, request)

			_, identityErr := identityRepository.Get(context.Background(), parts[0]+"@gmail.com")

			t.Log("Given the user with the same email exists in the external user table")
			t.Log("When the email is registering")
			t.Log("	Then the error is returned")
			require.ErrorIs(t, err, action.ErrUserAlreadyExists)

			t.Log("   And the identity is not created")
			require.ErrorIs(t, identityErr, repository.ErrIdentityNotFound)

		},
	)
}

func TestRegisterUser_Execute_Error(t *testing.T) {
	t.Parallel()
	t.Run(
		"invalid email host lookup", func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "@asdfadgadgadgadgadg.dsd",
				Password: gofakeit.Password(true, true, true, true, false, 20),
			}
			pair, err := register.Execute(ctx, request)

			t.Log("Given the identity is not registered")
			t.Log("When the identity is registering using invalid email host")
			t.Log("	Then error is returned")
			require.NotNil(t, err)
			require.Equal(t, "Email is not valid", errors.Hint(err))
			require.Empty(t, pair.AccessToken.Token.String)
		},
	)

	t.Run(
		"invalid email", func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "localhost",
				Password: gofakeit.Password(true, true, true, true, false, 20),
			}
			pair, err := register.Execute(ctx, request)

			t.Log("Given the identity is not registered")
			t.Log("When the identity is registering using invalid email host")
			t.Log("	Then error is returned")
			require.NotNil(t, err)
			require.Equal(t, "Email is not valid", errors.Hint(err))
			require.Empty(t, pair.AccessToken.Token.String)
		},
	)

	t.Run(
		"invalid password", func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			parts := strings.Split(gofakeit.Email(), "@")
			request := action.RegisterInput{
				Email:    parts[0] + "@google.com",
				Password: "",
			}
			pair, err := register.Execute(ctx, request)

			t.Log("Given the identity is not registered")
			t.Log("When the identity is registering using invalid email host")
			t.Log("	Then error is returned")
			require.NotNil(t, err)
			require.Equal(t, "Password is required", errors.Hint(err))
			require.Empty(t, pair.AccessToken.Token.String)
		},
	)
}
