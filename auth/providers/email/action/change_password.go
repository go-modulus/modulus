package action

import (
	"context"

	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordInput struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (i *ChangePasswordInput) Validate(ctx context.Context) error {
	return validator.ValidateStructWithContext(
		ctx,
		i,
		validation.Field(
			&i.OldPassword,
			validation.Required.Error("Old password is required"),
		),
		validation.Field(
			&i.NewPassword,
			passwordValidationRules...,
		),
	)
}

var ErrInvalidPassword = erruser.New(
	"invalid password",
	"The password you entered is incorrect. Please try again.",
)

type ChangePassword struct {
	credentialRepository repository.CredentialRepository
}

func NewChangePassword(credentialRepository repository.CredentialRepository) *ChangePassword {
	return &ChangePassword{
		credentialRepository: credentialRepository,
	}
}

func (c *ChangePassword) Execute(
	ctx context.Context,
	performerID uuid.UUID,
	input ChangePasswordInput,
) error {
	cred, err := c.credentialRepository.GetLast(ctx, performerID, string(repository.CredentialTypePassword))
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return errtrace.Wrap(ErrInvalidPassword)
		}
		return errtrace.Wrap(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(cred.Hash), []byte(input.OldPassword))
	if err != nil {
		return errtrace.Wrap(ErrInvalidPassword)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errtrace.Wrap(err)
	}

	err = c.credentialRepository.RemoveCredentials(ctx, performerID)
	if err != nil {
		return errtrace.Wrap(err)
	}

	_, err = c.credentialRepository.Create(
		ctx,
		performerID,
		string(hash),
		repository.CredentialTypePassword,
		nil,
	)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}
