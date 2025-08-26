package action_test

import (
	"context"
	"testing"

	"github.com/go-modulus/modulus/auth/providers/email/action"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestChangePassword_Execute(t *testing.T) {
	t.Parallel()
	t.Run(
		"Success - valid old password", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			oldPassword := "oldPassword123"
			newPassword := "newPassword456"

			// Create initial credential with old password
			oldHash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
			require.NoError(t, err)

			// Create an account and credential for testing
			accountID := uuid.Must(uuid.NewV4())
			account := authFixture.Account().ID(accountID).Create(t).GetEntity()
			authFixture.Credential().AccountID(accountID).Hash(string(oldHash)).Create(t).GetEntity()

			input := action.ChangePasswordInput{
				OldPassword: oldPassword,
				NewPassword: newPassword,
			}

			// Act
			err = changePassword.Execute(ctx, account.ID, input)

			authFixture.Credential().AccountID(accountID).CleanupAllOfAccount(t)
			// Assert
			require.NoError(t, err)

			// Verify the new password works
			newCred, err := credentialRepository.GetLast(ctx, account.ID, string(repository.CredentialTypePassword))
			require.NoError(t, err)

			err = bcrypt.CompareHashAndPassword([]byte(newCred.Hash), []byte(newPassword))
			assert.NoError(t, err, "New password should be properly hashed and stored")

			// Verify old password no longer works
			err = bcrypt.CompareHashAndPassword([]byte(newCred.Hash), []byte(oldPassword))
			assert.Error(t, err, "Old password should no longer work")
		},
	)

	t.Run(
		"Error - credential not found (no existing password)", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()

			// Create an account and credential for testing
			accountID := uuid.Must(uuid.NewV4())
			account := authFixture.Account().ID(accountID).Create(t).GetEntity()

			input := action.ChangePasswordInput{
				OldPassword: "somePassword123",
				NewPassword: "newPassword456",
			}

			// Act
			err := changePassword.Execute(ctx, account.ID, input)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, action.ErrInvalidPassword)
		},
	)

	t.Run(
		"Error - invalid old password", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()

			oldPassword := "oldPassword123"

			// Create initial credential with old password
			oldHash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
			require.NoError(t, err)

			// Create an account and credential for testing
			accountID := uuid.Must(uuid.NewV4())
			account := authFixture.Account().ID(accountID).Create(t).GetEntity()
			authFixture.Credential().AccountID(accountID).Hash(string(oldHash)).Create(t).GetEntity()

			input := action.ChangePasswordInput{
				OldPassword: "incorrectOldPassword", // Wrong password
				NewPassword: "newPassword456",
			}

			// Act
			err = changePassword.Execute(ctx, account.ID, input)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, action.ErrInvalidPassword)

			// Verify original password still works (credential wasn't changed)
			currentCred, err := credentialRepository.GetLast(ctx, account.ID, string(repository.CredentialTypePassword))
			require.NoError(t, err)

			err = bcrypt.CompareHashAndPassword([]byte(currentCred.Hash), []byte(oldPassword))
			assert.NoError(t, err, "Original password should still work")
		},
	)

	t.Run(
		"Error - account not found", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			nonExistentAccountID := uuid.Must(uuid.NewV4())

			input := action.ChangePasswordInput{
				OldPassword: "oldPassword123",
				NewPassword: "newPassword456",
			}

			// Act
			err := changePassword.Execute(ctx, nonExistentAccountID, input)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, action.ErrInvalidPassword)
		},
	)
}

func TestChangePasswordInput_Validate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run(
		"Success - valid input", func(t *testing.T) {
			t.Parallel()
			input := action.ChangePasswordInput{
				OldPassword: "oldPassword123",
				NewPassword: "newPassword456",
			}

			err := input.Validate(ctx)
			assert.NoError(t, err)
		},
	)

	t.Run(
		"Error - missing old password", func(t *testing.T) {
			t.Parallel()
			input := action.ChangePasswordInput{
				OldPassword: "",
				NewPassword: "newPassword456",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, errors.Hint(err), "Old password is required")
		},
	)

	t.Run(
		"Error - missing new password", func(t *testing.T) {
			t.Parallel()
			input := action.ChangePasswordInput{
				OldPassword: "oldPassword123",
				NewPassword: "",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, errors.Hint(err), "Password is required")
		},
	)

	t.Run(
		"Error - new password too short", func(t *testing.T) {
			t.Parallel()
			input := action.ChangePasswordInput{
				OldPassword: "oldPassword123",
				NewPassword: "123", // Too short
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, errors.Hint(err), "Password must be between 6 and 20 characters")
		},
	)

	t.Run(
		"Error - new password too long", func(t *testing.T) {
			t.Parallel()
			input := action.ChangePasswordInput{
				OldPassword: "oldPassword123",
				NewPassword: "thisPasswordIsTooLongAndShouldFailValidation", // Too long
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, errors.Hint(err), "Password must be between 6 and 20 characters")
		},
	)
}
