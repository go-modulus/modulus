package action

import (
	"context"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/gofrs/uuid"
	"log/slog"
)

var ErrUserAlreadyExists = erruser.New(
	"user already exists",
	"User with such email already exists. Please log in using another type of authentication.",
)

type User struct {
	ID       uuid.UUID
	Email    string
	UserInfo map[string]interface{}
}

type UserCreator interface {
	// CreateUser creates a new user with id and email.
	// It returns the created user.
	// Errors:
	// * ErrUserAlreadyExists - if the user already exists. In a case of this error it should return the existing user.
	CreateUser(ctx context.Context, user User) (User, error)
}

type DefaultUserCreator struct {
	logger *slog.Logger
}

func NewDefaultUserCreator(logger *slog.Logger) UserCreator {
	return &DefaultUserCreator{logger: logger}
}

func (c *DefaultUserCreator) CreateUser(ctx context.Context, user User) (User, error) {
	c.logger.Warn(
		`Override UserCreator with your own implementation.
In the main package create auth module as:
md := authEmail.OverrideUserCreator(authEmail.NewModule(), func(impl *UserCreatorImplementation) authEmailAction.UserCreator {
return impl
}`,
	)

	return user, nil
}
