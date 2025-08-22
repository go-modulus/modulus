package graphql

import (
	"context"
	"github.com/go-modulus/modulus/auth/graphql"
	"github.com/go-modulus/modulus/auth/repository"
	captchaAction "github.com/go-modulus/modulus/captcha/action"

	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/auth"
	mErrors "github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/erruser"

	"github.com/go-modulus/modulus/auth/providers/email/action"
)

var ErrWrongCredentials = erruser.New("wrong credentials", "Email or password is wrong")

type Resolver struct {
	register       *action.Register
	login          *action.Login
	resetPassword  *action.ResetPassword
	changePassword *action.ChangePassword
	checkCaptcha   *captchaAction.CheckCaptcha
}

func NewResolver(
	register *action.Register,
	login *action.Login,
	resetPassword *action.ResetPassword,
	changePassword *action.ChangePassword,
	checkCaptcha *captchaAction.CheckCaptcha,
) *Resolver {
	return &Resolver{
		register:       register,
		login:          login,
		resetPassword:  resetPassword,
		changePassword: changePassword,
		checkCaptcha:   checkCaptcha,
	}
}

func (r *Resolver) RegisterViaEmail(ctx context.Context, input RegisterViaEmailInput) (graphql.TokenPair, error) {
	err := r.checkCaptcha.Execute(input.Captcha)
	if err != nil {
		return graphql.TokenPair{}, errtrace.Wrap(err)
	}

	pair, err := r.register.Execute(
		ctx, action.RegisterInput{
			Email:    input.Email,
			Password: input.Password,
			UserInfo: input.UserInfo,
			Roles:    nil,
		},
	)
	if err != nil {
		return graphql.TokenPair{}, errtrace.Wrap(err)
	}
	auth.SendRefreshToken(ctx, pair.RefreshToken.Token.String)
	return r.transformPair(pair), nil
}

func (r *Resolver) LoginViaEmail(ctx context.Context, input LoginViaEmailInput) (graphql.TokenPair, error) {
	err := r.checkCaptcha.Execute(input.Captcha)
	if err != nil {
		return graphql.TokenPair{}, errtrace.Wrap(err)
	}
	pair, err := r.login.Execute(
		ctx, action.LoginInput{
			Email:    input.Email,
			Password: input.Password,
		},
	)
	if err != nil {
		if mErrors.Is(auth.ErrInvalidPassword, err) ||
			mErrors.Is(auth.ErrInvalidIdentity, err) {
			return graphql.TokenPair{}, ErrWrongCredentials
		}
		if mErrors.Is(auth.ErrIdentityIsBlocked, err) {
			return r.transformPair(pair), errtrace.Wrap(
				mErrors.WithHint(
					err,
					"Please contact the administrator. Your account is blocked.",
				),
			)
		}
	}

	auth.SendRefreshToken(ctx, pair.RefreshToken.Token.String)
	return r.transformPair(pair), err
}

func (r *Resolver) transformPair(pair auth.TokenPair) graphql.TokenPair {
	return graphql.TokenPair{
		AccessToken:  pair.AccessToken.Token.String,
		RefreshToken: pair.RefreshToken.Token.String,
	}
}

func (r *Resolver) RequestResetPassword(ctx context.Context, email string) (string, error) {
	_, err := r.resetPassword.Request(ctx, email)
	if err != nil {
		if mErrors.Is(err, repository.ErrIdentityNotFound) {
			return "", nil
		}
		return "", errtrace.Wrap(err)
	}

	return "", nil
}

func (r *Resolver) ConfirmResetPassword(
	ctx context.Context,
	input action.ConfirmResetPasswordInput,
) (string, error) {
	err := input.Validate(ctx)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	err = r.resetPassword.Confirm(ctx, input)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	return "", nil
}

func (r *Resolver) ChangePassword(ctx context.Context, input action.ChangePasswordInput) error {
	err := input.Validate(ctx)
	if err != nil {
		return errtrace.Wrap(err)
	}
	performerID := auth.GetPerformerID(ctx)
	err = r.changePassword.Execute(ctx, performerID, input)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}
