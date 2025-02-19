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
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (i *LoginUserInput) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(
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
		return validator.NewErrInvalidInputFromOzzo(ctx, err)
	}

	return nil
}

type LoginUser struct {
	passwordAuth *auth.PasswordAuthenticator
}

func NewLoginUser(passwordAuth *auth.PasswordAuthenticator) *LoginUser {
	return &LoginUser{passwordAuth: passwordAuth}
}

// Execute performs the login action by email and password.
// Returns a token pair of the access and refresh tokens if the login is successful.
// Errors:
// * github.com/go-modulus/modulus/auth.ErrIdentityIsBlocked - if the identity is blocked.
// * github.com/go-modulus/modulus/auth.ErrInvalidPassword - if the password is invalid.
// * Any error from the IdentityRepository.Get method (e.g. github.com/go-modulus/modulus/auth.ErrIdentityNotFound).
// * Any error from the CredentialRepository.GetLast method (e.g. github.com/go-modulus/modulus/auth.ErrCredentialNotFound).
func (l *LoginUser) Execute(ctx context.Context, input LoginUserInput) (TokenPair, error) {
	performer, err := l.passwordAuth.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

}
