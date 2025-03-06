package action

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type LoginUserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (i *LoginUserInput) Validate(ctx context.Context) error {
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

type LoginUser struct {
	passwordAuth   *auth.PasswordAuthenticator
	plainTokenAuth *auth.PlainTokenAuthenticator
}

func NewLoginUser(
	passwordAuth *auth.PasswordAuthenticator,
	tokenAuth *auth.PlainTokenAuthenticator,
) *LoginUser {
	return &LoginUser{
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
func (l *LoginUser) Execute(ctx context.Context, input LoginUserInput) (TokenPair, error) {
	err := input.Validate(ctx)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}
	// Authenticate the user with the given email and password.
	performer, err := l.passwordAuth.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

	// Issue a new pair of access and refresh tokens.
	pair, err := l.plainTokenAuth.IssueTokens(ctx, performer.IdentityID, nil)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

	return TokenPair{
		AccessToken:  pair.AccessToken.Token.String,
		RefreshToken: pair.RefreshToken.Token.String,
	}, nil
}
