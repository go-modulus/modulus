package storage

import (
	"braces.dev/errtrace"
	"context"
	"encoding/json"
	"errors"
	errors2 "github.com/go-modulus/modulus/auth/errors"
	errors3 "github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"
)

type CredentialType string

const CredentialTypePassword CredentialType = "password"
const CredentialTypeOTP CredentialType = "otp"

type DefaultRepository struct {
	queries *Queries
}

func NewDefaultRepository(db *pgxpool.Pool) *DefaultRepository {
	return &DefaultRepository{
		queries: New(db),
	}
}

func (r *DefaultRepository) MakeIdentity(
	ctx context.Context,
	identity string,
	UserID uuid.UUID,
	password string,
	AdditionalData map[string]interface{},
) error {
	_, err := r.GetIdentity(ctx, identity)
	if err == nil {
		return errors2.ErrIdentityExists
	} else if !errors.Is(err, errors2.ErrIdentityNotFound) {
		return errtrace.Wrap(err)
	}
	id := uuid.Must(uuid.NewV6())
	if AdditionalData == nil {
		AdditionalData = make(map[string]interface{})
	}
	dataVal, err := json.Marshal(AdditionalData)
	if err != nil {
		return errtrace.Wrap(err)
	}
	_, err = r.queries.CreateIdentity(
		ctx, CreateIdentityParams{
			ID:       id,
			Identity: identity,
			UserID:   UserID,
			Data:     dataVal,
		},
	)

	if err != nil {
		return errtrace.Wrap(errors3.WrapCause(errors2.ErrCannotCreateIdentity, err))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return errtrace.Wrap(errors3.WrapCause(errors2.ErrCannotHashPassword, err))
	}

	_, err = r.queries.CreateCredential(
		ctx, CreateCredentialParams{
			ID:             uuid.Must(uuid.NewV6()),
			IdentityID:     id,
			Type:           string(CredentialTypePassword),
			CredentialHash: string(hash),
			ExpiredAt:      null.Time{},
		},
	)

	if err != nil {
		return errtrace.Wrap(errors3.WrapCause(errors2.ErrCannotCreateCredential, err))
	}

	return nil
}

func (r *DefaultRepository) GetIdentity(
	ctx context.Context,
	identity string,
) (Identity, error) {
	res, err := r.queries.FindIdentity(ctx, identity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, errors2.ErrIdentityNotFound
		}
		return Identity{}, errtrace.Wrap(err)
	}
	return res, nil
}
