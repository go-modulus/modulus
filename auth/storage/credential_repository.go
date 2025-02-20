package storage

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
	"time"
)

type DefaultCredentialRepository struct {
	queries *Queries
}

func NewDefaultCredentialRepository(db *pgxpool.Pool) repository.CredentialRepository {
	return &DefaultCredentialRepository{
		queries: New(db),
	}
}

func (r *DefaultCredentialRepository) Create(
	ctx context.Context,
	identityID uuid.UUID,
	credHash string,
	credType string,
	expiredAt *time.Time,
) (repository.Credential, error) {
	expAt := null.TimeFromPtr(expiredAt)
	cred, err := r.queries.CreateCredential(
		ctx, CreateCredentialParams{
			IdentityID:     identityID,
			Type:           credType,
			CredentialHash: credHash,
			ExpiredAt:      expAt,
		},
	)

	if err != nil {
		return repository.Credential{}, errtrace.Wrap(errors.WrapCause(repository.ErrCannotCreateCredential, err))
	}

	return r.transform(cred), nil
}

func (r *DefaultCredentialRepository) transform(res Credential) repository.Credential {
	return repository.Credential{
		IdentityID:     res.IdentityID,
		CredentialHash: res.CredentialHash,
		Type:           res.Type,
		ExpiredAt:      res.ExpiredAt,
	}
}

func (r *DefaultCredentialRepository) GetLast(
	ctx context.Context,
	identityID uuid.UUID,
	credType string,
) (repository.Credential, error) {
	res, err := r.queries.FindLastCredentialOfType(
		ctx, FindLastCredentialOfTypeParams{
			IdentityID: identityID,
			Type:       credType,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Credential{}, repository.ErrCredentialNotFound
		}
		return repository.Credential{}, errtrace.Wrap(err)
	}
	return r.transform(res), nil
}
