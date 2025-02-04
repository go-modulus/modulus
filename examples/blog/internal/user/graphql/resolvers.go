package graphql

import (
	"blog/internal/user/action"
	"blog/internal/user/storage"
	"context"
)

type Resolver struct {
	register *action.RegisterUser
}

func NewResolver(
	register *action.RegisterUser,
) *Resolver {
	return &Resolver{
		register: register,
	}
}

func (r *Resolver) RegisterUser(ctx context.Context, input action.RegisterUserInput) (storage.User, error) {
	return r.register.Execute(ctx, input)
}
