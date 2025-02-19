package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"blog/internal/user/action"
	"blog/internal/user/storage"
	"context"
)

// RegisterUser is the resolver for the registerUser field.
func (r *mutationResolver) RegisterUser(ctx context.Context, input action.RegisterUserInput) (storage.User, error) {
	return r.userResolver.RegisterUser(ctx, input)
}

// LoginUser is the resolver for the loginUser field.
func (r *mutationResolver) LoginUser(ctx context.Context, input action.LoginUserInput) (action.TokenPair, error) {
	return r.userResolver.LoginUser(ctx, input)
}
