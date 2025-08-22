package action_test

import (
	"context"
	"github.com/go-modulus/modulus/auth/storage"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/providers/email/action"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin_Execute(t *testing.T) {
	t.Parallel()
	t.Run(
		"Success - valid email and password", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			email := strings.ToLower(gofakeit.FirstName() + "@gmail.com")
			password := "validPassword123"

			// Create account, identity, and credential
			accountID := uuid.Must(uuid.NewV4())
			identityID := uuid.Must(uuid.NewV4())

			authFixture.Account().ID(accountID).Create(t).GetEntity()
			authFixture.Identity().ID(identityID).AccountID(accountID).Identity(email).Create(t).GetEntity()

			// Hash password and create credential
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			require.NoError(t, err)
			authFixture.Credential().AccountID(accountID).Hash(string(passwordHash)).Create(t).GetEntity()

			input := action.LoginInput{
				Email:    email,
				Password: password,
			}

			// Act
			tokenPair, err := login.Execute(ctx, input)

			t.Log("When login")
			t.Log("   Then login is successful")
			require.NoError(t, err)

			t.Log("   Then access token is generated")
			assert.NotEmpty(t, tokenPair.AccessToken.Token.String)
			assert.Equal(t, accountID, tokenPair.AccessToken.AccountID)
			assert.Equal(t, identityID, tokenPair.AccessToken.IdentityID)

			t.Log("   Then refresh token is generated")
			assert.NotEmpty(t, tokenPair.RefreshToken.Token.String)
			assert.Equal(t, identityID, tokenPair.RefreshToken.IdentityID)

			t.Log("   Then tokens belong to same session")
			assert.Equal(t, tokenPair.AccessToken.SessionID, tokenPair.RefreshToken.SessionID)
		},
	)

	t.Run(
		"Success - email case insensitive", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			email := strings.ToLower(gofakeit.FirstName() + "@gmail.com")
			upperCaseEmail := strings.ToUpper(email)
			password := "validPassword123"

			// Create account, identity, and credential
			accountID := uuid.Must(uuid.NewV4())
			identityID := uuid.Must(uuid.NewV4())

			authFixture.Account().ID(accountID).Create(t).GetEntity()
			authFixture.Identity().ID(identityID).AccountID(accountID).Identity(email).Create(t).GetEntity()

			// Hash password and create credential
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			require.NoError(t, err)
			authFixture.Credential().AccountID(accountID).Hash(string(passwordHash)).Create(t).GetEntity()

			input := action.LoginInput{
				Email:    upperCaseEmail, // Using uppercase email
				Password: password,
			}

			// Act
			tokenPair, err := login.Execute(ctx, input)

			t.Log("When login")
			t.Log("   Then login with uppercase email is successful")
			require.NoError(t, err)

			t.Log("   Then tokens are generated correctly")
			assert.NotEmpty(t, tokenPair.AccessToken.Token.String)
			assert.Equal(t, accountID, tokenPair.AccessToken.AccountID)
			assert.Equal(t, identityID, tokenPair.AccessToken.IdentityID)
		},
	)

	t.Run(
		"Error - invalid email format", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			input := action.LoginInput{
				Email:    "invalid-email-format",
				Password: "validPassword123",
			}

			// Act
			_, err := login.Execute(ctx, input)

			t.Log("When login")
			t.Log("   Then validation error is returned")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "LoginInput.email")
		},
	)

	t.Run(
		"Error - identity not found", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			input := action.LoginInput{
				Email:    gofakeit.FirstName() + "@gmail.com", // Non-existent email
				Password: "validPassword123",
			}

			// Act
			_, err := login.Execute(ctx, input)

			t.Log("When login")
			t.Log("   Then authentication error is returned")
			require.Error(t, err)
			assert.ErrorIs(t, err, auth.ErrInvalidIdentity)
		},
	)

	t.Run(
		"Error - invalid password", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			email := strings.ToLower(gofakeit.FirstName() + "@gmail.com")
			correctPassword := "correctPassword123"
			wrongPassword := "wrongPassword123"

			// Create account, identity, and credential
			accountID := uuid.Must(uuid.NewV4())
			identityID := uuid.Must(uuid.NewV4())

			authFixture.Account().ID(accountID).Create(t).GetEntity()
			authFixture.Identity().ID(identityID).AccountID(accountID).Identity(email).Create(t).GetEntity()

			// Hash correct password and create credential
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
			require.NoError(t, err)
			authFixture.Credential().AccountID(accountID).Hash(string(passwordHash)).Create(t).GetEntity()

			input := action.LoginInput{
				Email:    email,
				Password: wrongPassword, // Wrong password
			}

			// Act
			_, err = login.Execute(ctx, input)

			t.Log("When login")
			t.Log("   Then invalid password error is returned")
			require.Error(t, err)
			assert.ErrorIs(t, err, auth.ErrInvalidPassword)
		},
	)

	t.Run(
		"Error - blocked identity", func(t *testing.T) {
			t.Parallel()
			// Arrange
			ctx := context.Background()
			email := strings.ToLower(gofakeit.FirstName() + "@gmail.com")
			password := "validPassword123"

			// Create account, blocked identity, and credential
			accountID := uuid.Must(uuid.NewV4())
			identityID := uuid.Must(uuid.NewV4())

			authFixture.Account().ID(accountID).Create(t).GetEntity()
			authFixture.Identity().ID(identityID).AccountID(accountID).Identity(email).Status(storage.IdentityStatusBlocked).Create(t).GetEntity()

			// Hash password and create credential
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			require.NoError(t, err)
			authFixture.Credential().AccountID(accountID).Hash(string(passwordHash)).Create(t).GetEntity()

			input := action.LoginInput{
				Email:    email,
				Password: password,
			}

			// Act
			_, err = login.Execute(ctx, input)

			t.Log("When login")
			t.Log("   Then identity blocked error is returned")
			require.Error(t, err)
			assert.ErrorIs(t, err, auth.ErrIdentityIsBlocked)
		},
	)
}

func TestLoginInput_Validate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run(
		"Success - valid input", func(t *testing.T) {
			t.Parallel()
			input := action.LoginInput{
				Email:    gofakeit.FirstName() + "@gmail.com",
				Password: "validPassword123",
			}

			err := input.Validate(ctx)
			assert.NoError(t, err)
		},
	)

	t.Run(
		"Error - invalid email", func(t *testing.T) {
			t.Parallel()
			input := action.LoginInput{
				Email:    "invalid-email",
				Password: "validPassword123",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "LoginInput.email")
		},
	)

	t.Run(
		"Error - missing email", func(t *testing.T) {
			t.Parallel()
			input := action.LoginInput{
				Email:    "",
				Password: "validPassword123",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "LoginInput.email")
		},
	)

	t.Run(
		"Error - missing password", func(t *testing.T) {
			t.Parallel()
			input := action.LoginInput{
				Email:    gofakeit.FirstName() + "@gmail.com",
				Password: "",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "LoginInput.password")
		},
	)

	t.Run(
		"Error - password too short", func(t *testing.T) {
			t.Parallel()
			input := action.LoginInput{
				Email:    gofakeit.FirstName() + "@gmail.com",
				Password: "123",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "LoginInput.password")
		},
	)

	t.Run(
		"Error - password too long", func(t *testing.T) {
			t.Parallel()
			input := action.LoginInput{
				Email:    gofakeit.FirstName() + "@gmail.com",
				Password: "thisPasswordIsTooLongAndShouldFailValidation",
			}

			err := input.Validate(ctx)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "LoginInput.password")
		},
	)
}
