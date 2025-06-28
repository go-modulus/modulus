package action

import (
	"context"
	"fmt"
	"strings"

	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
)

const DefaultUserRole = "user"

var ErrEmailAlreadyExists = erruser.New(
	"email already exists",
	"Email already exists. Please log in using your password.",
)

var ErrCannotLogin = erruser.New(
	"cannot login",
	"Cannot log in automatically after registration. Please, try to log in yourself.",
)

type RegisterInput struct {
	Email                      string
	Password                   string
	UserInfo                   map[string]interface{}
	SubscribeToMarketingEmails bool
	Roles                      []string
}

var passwordValidationRules = []validation.Rule{
	validation.Required.Error("Password is required"),
	validation.Length(6, 20).Error("Password must be between 6 and 20 characters"),
}

func (i *RegisterInput) Validate(ctx context.Context) error {
	err := validator.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(
			&i.Email,
			validation.Required.Error("Email is required"),
			is.Email.Error("Email is not valid"),
		),
		validation.Field(
			&i.Password,
			passwordValidationRules...,
		),
	)

	if err != nil {
		return err
	}

	return nil
}

type Register struct {
	passwordAuth       *auth.PasswordAuthenticator
	identityRepository repository.IdentityRepository
	accountRepository  repository.AccountRepository
	userCreator        UserCreator
	login              *Login
}

func NewRegister(
	passwordAuth *auth.PasswordAuthenticator,
	identityRepository repository.IdentityRepository,
	accountRepository repository.AccountRepository,
	userCreator UserCreator,
	login *Login,
) *Register {
	return &Register{
		passwordAuth:       passwordAuth,
		identityRepository: identityRepository,
		accountRepository:  accountRepository,
		userCreator:        userCreator,
		login:              login,
	}
}

// Execute performs the register action by email and password.
// It returns the registered user.
//
// Errors:
// * ErrEmailAlreadyExists - if the user already exists.
// * ErrUserAlreadyExists - if cannot log in automatically.
// * ErrCannotLogin - if cannot log in automatically.
func (r *Register) Execute(ctx context.Context, input RegisterInput) (auth.TokenPair, error) {
	err := input.Validate(context.Background())
	if err != nil {
		return auth.TokenPair{}, err
	}

	input.Email = strings.ToLower(input.Email)
	identityStr := input.Email

	if len(input.Roles) == 0 {
		input.Roles = []string{DefaultUserRole}
	}

	// Register new account with the email as identity
	account, err := r.passwordAuth.Register(
		ctx,
		identityStr,
		input.Password,
		repository.IdentityTypeEmail,
		// the authenticated user role that will be used in the future
		input.Roles,
		input.UserInfo,
	)
	if err != nil {
		if errors.Is(err, repository.ErrIdentityExists) {
			return auth.TokenPair{}, ErrEmailAlreadyExists
		}
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	// Create a new user.
	existingUser, err := r.userCreator.CreateUser(
		ctx, User{
			ID:       account.ID,
			Email:    input.Email,
			UserInfo: input.UserInfo,
		},
	)
	err = r.processUserCreationError(ctx, err, account.ID, existingUser)
	if err != nil {
		return auth.TokenPair{}, err
	}

	// Issue a new pair of access and refresh tokens.
	pair, err := r.login.Execute(
		ctx, LoginInput{
			Email:    input.Email,
			Password: input.Password,
		},
	)

	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(erruser.WithCause(ErrCannotLogin, err))
	}

	return pair, nil
}

func (r *Register) processUserCreationError(
	ctx context.Context,
	err error,
	accountID uuid.UUID,
	existingUser User,
) error {
	if err == nil {
		return nil
	}
	err2 := r.accountRepository.RemoveAccount(ctx, accountID)
	if err2 != nil {
		return errtrace.Wrap(err2)
	}
	// If the user already exists with such an email, and the contract of interface is met,
	// try to help the user to log in using another type of authentication.
	if errors.Is(err, ErrUserAlreadyExists) && existingUser.ID != uuid.Nil {
		identities, err2 := r.identityRepository.GetByAccountID(ctx, existingUser.ID)
		if err2 != nil {
			return errtrace.Wrap(err2)
		}
		if len(identities) > 0 {
			return errors.WithHint(
				ErrUserAlreadyExists,
				fmt.Sprintf(
					"Please log in using another type of authentication. You have registered using %s.",
					identities[0].Type,
				),
			)
		} else {
			return erruser.WithCause(
				ErrUserAlreadyExists,
				errors.WithMeta(
					errors.New(
						"inconsistent state",
					),
					"user.email", existingUser.Email,
				),
			)
		}
	}
	return errtrace.Wrap(err)
}
