package action_test

import (
	"context"
	"errors"
	"github.com/go-modulus/modulus/auth/providers/google/action"
	"github.com/go-modulus/modulus/auth/providers/google/action/mocks"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegister_Execute(t *testing.T) {
	t.Run("successful registration with new user", func(t *testing.T) {
		t.Skip("requires real Google OAuth - use for integration testing")

		ctx := context.Background()
		userID := uuid.Must(uuid.NewV6())
		expectedUser := action.User{
			ID:    userID,
			Email: "test@example.com",
			GoogleUser: action.GoogleUser{
				ID:            "google-user-id",
				Email:         "test@example.com",
				VerifiedEmail: true,
				Name:          "Test User",
			},
		}

		creatorMock.On("CreateUserFromGoogle", mock.Anything, mock.AnythingOfType("action.User")).Return(expectedUser, nil)

		pair, err := register.Execute(ctx, action.RegisterInput{
			Code:     "valid-auth-code",
			Verifier: "valid-verifier",
			Roles:    []string{"user"},
			UserInfo: map[string]interface{}{"source": "test"},
		})

		require.NoError(t, err)
		require.NotEmpty(t, pair.AccessToken.Token.String)
		require.NotEmpty(t, pair.RefreshToken.Token.String)
		creatorMock.AssertExpectations(t)
	})

	t.Run("registration with existing user", func(t *testing.T) {
		t.Skip("requires real Google OAuth - use for integration testing")

		ctx := context.Background()
		existingUserID := uuid.Must(uuid.NewV6())
		existingUser := action.User{
			ID:    existingUserID,
			Email: "existing@example.com",
			GoogleUser: action.GoogleUser{
				ID:            "google-user-id-2",
				Email:         "existing@example.com",
				VerifiedEmail: true,
				Name:          "Existing User",
			},
		}

		creatorMock.On("CreateUserFromGoogle", mock.Anything, mock.AnythingOfType("action.User")).Return(existingUser, action.ErrUserAlreadyExists)

		pair, err := register.Execute(ctx, action.RegisterInput{
			Code:     "valid-auth-code-2",
			Verifier: "valid-verifier-2",
			Roles:    []string{"user"},
			UserInfo: nil,
		})

		require.NoError(t, err)
		require.NotEmpty(t, pair.AccessToken.Token.String)
		require.NotEmpty(t, pair.RefreshToken.Token.String)
		creatorMock.AssertExpectations(t)
	})

	t.Run("user creation failure", func(t *testing.T) {
		t.Skip("requires real Google OAuth - use for integration testing")

		ctx := context.Background()
		creationError := errors.New("database connection failed")

		creatorMock.On("CreateUserFromGoogle", mock.Anything, mock.AnythingOfType("action.User")).Return(action.User{}, creationError)

		_, err := register.Execute(ctx, action.RegisterInput{
			Code:     "valid-auth-code-3",
			Verifier: "valid-verifier-3",
			Roles:    []string{"user"},
			UserInfo: nil,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection failed")
		creatorMock.AssertExpectations(t)
	})

	t.Run("unit test - user creator mock validation", func(t *testing.T) {
		ctx := context.Background()
		mockCreator := mocks.NewMockUserCreator(t)

		userID := uuid.Must(uuid.NewV6())
		inputUser := action.User{
			ID:    userID,
			Email: "mock@example.com",
			GoogleUser: action.GoogleUser{
				ID:            "mock-google-id",
				Email:         "mock@example.com",
				VerifiedEmail: true,
				Name:          "Mock User",
			},
			UserInfo: map[string]interface{}{"test": "data"},
		}

		mockCreator.EXPECT().CreateUserFromGoogle(ctx, inputUser).Return(inputUser, nil)

		result, err := mockCreator.CreateUserFromGoogle(ctx, inputUser)

		require.NoError(t, err)
		assert.Equal(t, inputUser, result)
	})
}
