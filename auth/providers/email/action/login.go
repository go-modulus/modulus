package action

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"strings"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i *LoginInput) Validate(ctx context.Context) error {
	err := validator.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(
			&i.Email,
			validation.Required.Error("Email is required"),
			is.Email.Error("Email is not valid"),
		),
		validation.Field(
			&i.Password,
			validation.Required.Error("Password is required"),
			validation.Length(6, 20).Error("Password must be between 6 and 20 characters"),
		),
	)

	if err != nil {
		return err
	}

	return nil
}

type Login struct {
	passwordAuth   *auth.PasswordAuthenticator
	plainTokenAuth *auth.PlainTokenAuthenticator
}

func NewLogin(
	passwordAuth *auth.PasswordAuthenticator,
	tokenAuth *auth.PlainTokenAuthenticator,
) *Login {
	return &Login{
		passwordAuth:   passwordAuth,
		plainTokenAuth: tokenAuth,
	}
}

// Execute performs the login action by email and password.
// Returns a token pair of the access and refresh tokens if the login is successful.
// Errors:
// * github.com/go-modulus/modulus/auth.ErrIdentityIsBlocked - if the identity is blocked.
// * github.com/go-modulus/modulus/auth.ErrInvalidPassword - if the password is invalid.
// * github.com/go-modulus/modulus/auth.ErrInvalidIdentity - if identity is not found in the repository.
func (l *Login) Execute(ctx context.Context, input LoginInput) (auth.TokenPair, error) {
	err := input.Validate(context.Background())
	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(err)
	}
	identityStr := strings.ToLower(input.Email)
	// Authenticate the user with the given email and password.
	performer, err := l.passwordAuth.Authenticate(ctx, identityStr, input.Password)
	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	// Issue a new pair of access and refresh tokens.
	pair, err := l.plainTokenAuth.IssueTokens(ctx, performer.IdentityID, nil)
	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	return pair, nil
}
