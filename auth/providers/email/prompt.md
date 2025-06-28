# Email Provider for the Auth Module
This is a prompt with actions to implement after the module installation.

1. If module `user` is not created in the `internal/user` folder, create it using the command `make module-create`.
2. If module `mail` is not created in the `internal/mail` folder, create it using the command `make module-create`.
3. In the user module create a file `internal/user/action/register_user.go` with the content like this:
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
			IsSubscribed: input.SubscribeToMarketingEmails,
		},
	)
	if err != nil {
		return storage.User{}, errtrace.Wrap(errsys.WithCause(ErrCannotRegisterUser, err))
	}

	return user, nil
}

// CreateUser creates a new user with id and email.
// It returns the created user.
// Errors:
// * i10x/internal/auth/email/action.ErrUserAlreadyExists - if the user already exists. In a case of this error it should return the existing user.
func (r *RegisterUser) CreateUser(ctx context.Context, user action.User) (action.User, error) {
	u, err := r.userDb.FindUserByEmail(ctx, user.Email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return action.User{}, errtrace.Wrap(err)
		}
	} else {
		return action.User{
			ID:    u.ID,
			Email: u.Email,
		}, ErrUserAlreadyExists
	}
	u, err = r.Execute(
		ctx, RegisterUserInput{
			ID:                         user.ID,
			Email:                      user.Email,
			Name:                       user.UserInfo["name"],
			IsEmailVerified:            false,
		},
	)
	if err != nil {
		return action.User{}, err
	}
	return action.User{
		ID:                          u.ID,
		Email:                       u.Email,
	}, nil
}
```
It should:
* implement the `github.com/go-modulus/modulus/auth/providers/email/action.UserCreator` interface.
* return `github.com/go-modulus/modulus/auth/providers/email/action.ErrUserAlreadyExists` error if the user with the same email already exists in the database. In this case it should return the existing user.
* save the user in the database with the `IsEmailVerified` field set to false.
* transform the `github.com/go-modulus/modulus/auth/providers/email/action.User` to the `github.com/go-modulus/modulus/storage.User` type.
4. In the user module create a file `internal/user/action/verify_email.go`. It should check if the in the user table gotten email is present and verified. It should implement the `github.com/go-modulus/modulus/auth/providers/email/action.VerifiedEmailChecker` interface. It should return `true` if the email is verified and `false` if it is not verified or does not exist in the database. Also, it returns the ID of found user. The SQL code can be the following:
```sql
SELECT u.id
FROM "user"."user" as u
         JOIN "user".email_verification ev on u.id = ev.user_id AND lower(u.email) = lower(ev.email)
WHERE lower(u.email) = lower($1::text)
  AND is_verified = true`
```
5. In the mail module create a file `internal/mail/action/reset_password_email_sender.go`. It should implement the `github.com/go-modulus/modulus/auth/providers/email/action.ResetPasswordEmailSender` interface. It should send an email with a link to reset the password. The link should contain a token that is valid for 1 hour.
6. In the `cmd/console/main.go` file change the constructor `authEmail.NewModule to 
```go
    authEmail.NewModule().WithOptions(
			authEmail.OverrideMailSender[*mailAction.ResetPasswordEmail],
			authEmail.OverrideVerifiedEmailChecker[*userAction.EmailVerifiedChecker],
			authEmail.OverrideUserCreator[*userAction.RegisterUser],
		).Module,
```
Write as a type in generic the created action.
