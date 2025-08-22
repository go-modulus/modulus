package action

import (
	"braces.dev/errtrace"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

var ErrResetPasswordTokenNotFound = erruser.New(
	"reset password token not found",
	"The reset password link is invalid or has expired.",
)
var ErrUserNotFound = errsys.New(
	"user not found",
	"The user with the provided email address does not exist.",
)

type ConfirmResetPasswordInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (i *ConfirmResetPasswordInput) Validate(ctx context.Context) error {
	return validator.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(
			&i.Token,
			validation.Required.Error("Token is required"),
		),
		validation.Field(
			&i.Password,
			passwordValidationRules...,
		),
	)
}

func GenerateToken() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errtrace.Wrap(err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

type ResetPasswordConfig struct {
	FrontendHost string `env:"FRONTEND_HOST" envDefault:"http://localhost:8001"`
}

type MailSender interface {
	ResetPasswordEmail(
		ctx context.Context,
		to string,
		resetPasswordLink string,
	) error
}

type DefaultMailSender struct {
	logger *slog.Logger
}

func (d *DefaultMailSender) ResetPasswordEmail(
	ctx context.Context,
	to string,
	resetPasswordLink string,
) error {
	// Implementation of sending email
	d.logger.Info("Sending reset password email", "to", to, "link", resetPasswordLink)
	return nil
}

func NewDefaultMailSender(logger *slog.Logger) MailSender {
	return &DefaultMailSender{
		logger: logger,
	}
}

type VerifiedEmailChecker interface {
	// FindUserIDByVerifiedEmail checks if the email is verified and returns the user ID.
	// If the email is not verified or not found, it returns ErrUserNotFound.
	FindUserIDByVerifiedEmail(ctx context.Context, email string) (uuid.UUID, error)
}
type DefaultVerifiedEmailChecker struct {
}

func (d *DefaultVerifiedEmailChecker) FindUserIDByVerifiedEmail(ctx context.Context, email string) (uuid.UUID, error) {
	return uuid.Nil, ErrUserNotFound
}

func NewDefaultVerifiedEmailChecker() VerifiedEmailChecker {
	return &DefaultVerifiedEmailChecker{}
}

type ResetPassword struct {
	config        ResetPasswordConfig
	identity      repository.IdentityRepository
	account       repository.AccountRepository
	resetPassword repository.ResetPasswordRequestRepository
	mailer        MailSender
	credentials   repository.CredentialRepository
	emailChecker  VerifiedEmailChecker
}

func NewResetPassword(
	config ResetPasswordConfig,
	identity repository.IdentityRepository,
	account repository.AccountRepository,
	resetPassword repository.ResetPasswordRequestRepository,
	mailer MailSender,
	credentials repository.CredentialRepository,
	emailChecker VerifiedEmailChecker,
) *ResetPassword {
	return &ResetPassword{
		config:        config,
		identity:      identity,
		account:       account,
		resetPassword: resetPassword,
		mailer:        mailer,
		credentials:   credentials,
		emailChecker:  emailChecker,
	}
}

func (r *ResetPassword) Request(ctx context.Context, email string) (repository.ResetPasswordRequest, error) {
	identity, err := r.getIdentity(ctx, email)
	if err != nil {
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}

	resetPassword, err := r.resetPassword.GetActiveRequest(ctx, identity.AccountID)
	if err != nil {
		if errors.Is(err, repository.ErrResetPasswordRequestNotFound) {
			return r.sendNew(ctx, identity)
		}
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}

	if !resetPassword.IsAlive() {
		err = r.resetPassword.ExpireRequest(ctx, resetPassword.ID)
		if err != nil {
			return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
		}
		return r.sendNew(ctx, identity)
	}

	if resetPassword.CanBeResent() {
		return resetPassword, r.sendEmail(ctx, identity.Identity, resetPassword)
	}

	return resetPassword, nil
}

func (r *ResetPassword) getIdentity(ctx context.Context, email string) (repository.Identity, error) {
	identity, err := r.identity.Get(ctx, email)
	if err == nil {
		return identity, nil
	}

	if !errors.Is(err, repository.ErrIdentityNotFound) {
		return repository.Identity{}, errtrace.Wrap(err)
	}

	userID, err := r.emailChecker.FindUserIDByVerifiedEmail(ctx, email)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return repository.Identity{}, repository.ErrIdentityNotFound
		}
		return repository.Identity{}, errtrace.Wrap(err)
	}

	account, err := r.account.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return repository.Identity{}, repository.ErrIdentityNotFound
		}
		return repository.Identity{}, errtrace.Wrap(err)
	}

	// if the user is authenticated using another provider or added by an admin with account creation and the email is verified
	// create a new identity for him and reset the password
	return r.identity.Create(
		ctx,
		email,
		account.ID,
		repository.IdentityTypeEmail,
		map[string]interface{}{
			"source": "reset password form",
		},
	)
}

func (r *ResetPassword) sendNew(ctx context.Context, identity repository.Identity) (
	repository.ResetPasswordRequest,
	error,
) {
	token, err := GenerateToken()
	if err != nil {
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}

	resetPassword, err := r.resetPassword.CreateResetPassword(
		ctx,
		uuid.Must(uuid.NewV7()),
		identity.AccountID,
		token,
	)
	if err != nil {
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}

	err = r.sendEmail(ctx, identity.Identity, resetPassword)
	if err != nil {
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}

	return resetPassword, nil
}

func (r *ResetPassword) sendEmail(
	ctx context.Context,
	email string,
	resetPassword repository.ResetPasswordRequest,
) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", r.config.FrontendHost, resetPassword.Token)

	err := r.mailer.ResetPasswordEmail(ctx, email, link)
	if err != nil {
		return errtrace.Wrap(err)
	}

	err = r.resetPassword.UpdateLastSent(ctx, resetPassword.ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *ResetPassword) Confirm(ctx context.Context, input ConfirmResetPasswordInput) error {
	resetPassword, err := r.resetPassword.GetResetPasswordByToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, repository.ErrResetPasswordRequestNotFound) {
			return ErrResetPasswordTokenNotFound
		}
		return errtrace.Wrap(err)
	}

	err = r.credentials.RemoveCredentials(ctx, resetPassword.AccountID)
	if err != nil {
		return errtrace.Wrap(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return errtrace.Wrap(err)
	}

	_, err = r.credentials.Create(
		ctx,
		resetPassword.AccountID,
		string(hash),
		repository.CredentialTypePassword,
		nil,
	)
	if err != nil {
		return errtrace.Wrap(err)
	}

	err = r.resetPassword.UseResetPassword(ctx, resetPassword.ID)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return nil
}
