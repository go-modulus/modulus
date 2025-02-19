package graphql

import (
	"blog/internal/user/action"
	"blog/internal/user/storage"
	"context"
)

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
	return r.login.Execute(ctx, input)
}
