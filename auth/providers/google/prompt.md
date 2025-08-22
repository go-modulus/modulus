# Google Provider for the Auth Module
This is a prompt with actions to implement after the module installation.

1. If there is no a module folder `internal/user`, create it using the command `mtools module create --path=internal/user --package=user --name="User managment module"`.
2. If the file `internal/user/action/register_user.go` is not exist create it  with the content like this:
```go
package action

import (
	"braces.dev/errtrace"
	"context"
	"errors"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/go-modulus/modulus/auth/providers/email/action"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

)

var ErrCannotRegisterUser = errsys.New(
	"cannot register user",
	"There are some issues in the registration process. Try to register again later.",
)

type RegisterUserInput struct {
	ID                         uuid.UUID
	Email                      string
	Name                       string
	IsEmailVerified            bool
}

func (i *RegisterUserInput) Validate(ctx context.Context) error {
	err := validator.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(
			&i.Email,
			validation.Required.Error("Email is required"),
			is.Email.Error("Email is not valid"),
		),
	)

	if err != nil {
		return err
	}

	return nil
}

type RegisterUser struct {
	userDb              *storage.Queries
}

func NewRegisterUser(
	userDb *storage.Queries,
) *RegisterUser {
	return &RegisterUser{
		userDb:              userDb,
	}
}

// Execute performs the register action by email and password.
// It returns the registered user.
//
// Errors:
// * ErrUserAlreadyExists - if the user already exists.
// * ErrInvalidEmail - if the email is invalid.
func (r *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (storage.User, error) {
	err := input.Validate(context.Background())
	if err != nil {
		return storage.User{}, err
	}

	email := input.Email

	_, err = r.userDb.FindUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return storage.User{}, errtrace.Wrap(err)
		}
	} else {
		return storage.User{}, action.ErrUserAlreadyExists
	}

	if input.ID == uuid.Nil {
		input.ID = uuid.Must(uuid.NewV6())
	}
	user, err := r.userDb.RegisterUser(
		ctx, storage.RegisterUserParams{
			ID:           input.ID,
			Email:        email,
			Name:         input.Name,
		},
	)
	if err != nil {
		return storage.User{}, errtrace.Wrap(errsys.WithCause(ErrCannotRegisterUser, err))
	}

	return user, nil
}

// CreateUserFromGoogle creates a new user with id and email.
// It returns the created user.
// Errors:
// * i10x/internal/auth/google/action.ErrUserAlreadyExists - if the user already exists. In a case of this error it should return the existing user.
func (r *RegisterUser) CreateUserFromGoogle(ctx context.Context, user action2.User) (action2.User, error) {
	u, err := r.userDb.FindUserByEmail(ctx, user.Email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return action2.User{}, errtrace.Wrap(err)
		}
	} else {
		return action2.User{
			ID:                          u.ID,
			Email:                       u.Email,
		}, ErrUserAlreadyExists
	}
	u, err = r.Execute(
		ctx, RegisterUserInput{
			ID:                         user.ID,
			Email:                      user.Email,
			Name:                       user.GoogleUser.Name,
			IsEmailVerified:            true,
		},
	)
	if err != nil {
		return action2.User{}, err
	}
	return action2.User{
		ID:                          u.ID,
		Email:                       u.Email,
	}, nil
}
```
It should:
* implement the `github.com/go-modulus/modulus/auth/providers/google/action.UserCreator` interface.
* return `github.com/go-modulus/modulus/auth/providers/google/action.ErrUserAlreadyExists` error if the user with the same email already exists in the database. In this case it should return the existing user.
* save the user in the database with the `IsEmailVerified` field set to true.

3. In the `cmd/console/main.go` file change the constructor `authGoogle.NewModule to
```go
authGoogle.OverrideUserCreator[*userAction.RegisterUser](
    authGoogle.NewModule(),
)
```
Write as a type in generic the created action.
4. If there is no `internal/graphql/resolver/auth.resolvers.go` file make it running `make graphql-generate`.
5. In the `internal/graphql/resolver/auth.resolvers.go` edit all generated resolvers adding a call to the appropriate resolvers from the `github.com/go-modulus/modulus/auth/providers/google/graphql.Resolver` struct. For example, change
```go
func (r *mutationResolver) RegisterViaEmail(ctx context.Context, input graphql1.RegisterViaGoogleInput) (graphql2.TokenPair, error) {
	panic(fmt.Errorf("not implemented: RegisterViaEmail - registerViaEmail"))
}
````
with
```go
func (r *mutationResolver) RegisterViaGoogle(ctx context.Context, input graphql1.RegisterViaGoogleInput) (graphql2.TokenPair, error) {
    return r.authGoogleResolver.RegisterViaGoogle(ctx, input)
}
```
6. In the `cmd/console/main.go` file change the constructor `http.NewModule` to
```go
http.OverrideErrorPipeline(
			http.OverrideMiddlewarePipeline(
				http.NewModule(),
				func(
					logger *slog.Logger,
					authMd *auth.Middleware,
				) *http.Pipeline {
					return &http.Pipeline{
						Middlewares: []http.Middleware{
							middleware.RequestID,
							middleware.IP,
							middleware.UserAgent,
							middleware.NewLogger(logger),
							authMd.HttpMiddleware(),
						},
					}
				},
			),
			func(
				logger *slog.Logger,
				loggerConfig errhttp.ErrorLoggerConfig,
			) *errhttp.ErrorPipeline {
				defPipeline := errhttp.NewDefaultErrorPipeline(logger, loggerConfig)
				defPipeline.Processors = append(defPipeline.Processors, auth.AddHttpCode())
				return defPipeline
			},
		),
```
If this change is not already done.