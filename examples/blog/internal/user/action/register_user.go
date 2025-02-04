package action

import (
	"blog/internal/user/storage"
	"braces.dev/errtrace"
	"context"
	"errors"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrUserAlreadyExists = erruser.New("user already exists", "User already exists. Please login or use another email.")

type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

func (i *RegisterUserInput) Validate(ctx context.Context) error {
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
		validation.Field(
			&i.Name,
			validation.Required.Error("Name is required"),
			is.Alpha.Error("Name must contain only letters"),
		),
	)

	if err != nil {
		return validator.NewErrInvalidInputFromOzzo(ctx, err)
	}

	return nil
}

type RegisterUser struct {
	userDb *storage.Queries
}

func NewRegisterUser(userDb *storage.Queries) *RegisterUser {
	return &RegisterUser{userDb: userDb}
}

func (r *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (storage.User, error) {
	err := input.Validate(context.Background())
	if err != nil {
		return storage.User{}, err
	}

	_, err = r.userDb.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return storage.User{}, errtrace.Wrap(err)
		}
	} else {
		return storage.User{}, ErrUserAlreadyExists
	}
	user, err := r.userDb.RegisterUser(
		ctx, storage.RegisterUserParams{
			ID:    uuid.Must(uuid.NewV6()),
			Email: input.Email,
			Name:  input.Name,
		},
	)
	if err != nil {
		return storage.User{}, errtrace.Wrap(err)
	}
	return user, nil
}
