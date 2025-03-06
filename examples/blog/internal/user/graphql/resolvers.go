package graphql

import (
	"blog/internal/user/action"
	"blog/internal/user/storage"
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth"
	mErrors "github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/erruser"
)

var ErrWrongCredentials = erruser.New("wrong credentials", "Email or password is wrong")

type Resolver struct {
	register *action.RegisterUser
	login    *action.LoginUser
}

func NewResolver(
	register *action.RegisterUser,
	login *action.LoginUser,
) *Resolver {
	return &Resolver{
		register: register,
		login:    login,
	}
}

func (r *Resolver) RegisterUser(ctx context.Context, input action.RegisterUserInput) (storage.User, error) {
	return r.register.Execute(ctx, input)
}

func (r *Resolver) LoginUser(ctx context.Context, input action.LoginUserInput) (action.TokenPair, error) {
	err := input.Validate(ctx)
	if err != nil {
		return action.TokenPair{}, errtrace.Wrap(err)
	}
	token, err := r.login.Execute(ctx, input)
	if err != nil {
		if mErrors.Is(auth.ErrInvalidPassword, err) ||
			mErrors.Is(auth.ErrInvalidIdentity, err) {
			return action.TokenPair{}, ErrWrongCredentials
		}
		if mErrors.Is(auth.ErrIdentityIsBlocked, err) {
			return token, errtrace.Wrap(
				mErrors.WithHint(
					err,
					"Please contact the administrator. Your account is blocked.",
				),
			)
		}
	}

	return token, err
}
