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
	accountID uuid.UUID,
	credentialHash string,
	credType repository.CredentialType,
	expiredAt *time.Time,
) (repository.Credential, error) {
	expAt := null.TimeFromPtr(expiredAt)
	cred, err := r.queries.CreateCredential(
		ctx, CreateCredentialParams{
			AccountID: accountID,
			Type:      string(credType),
			Hash:      credentialHash,
			ExpiredAt: expAt,
		},
	)

	if err != nil {
		return repository.Credential{}, errtrace.Wrap(errors.WithCause(repository.ErrCannotCreateCredential, err))
	}

	return r.transform(cred), nil
}

func (r *DefaultCredentialRepository) RemoveCredentials(ctx context.Context, accountID uuid.UUID) error {
	return errtrace.Wrap(r.queries.RemoveCredentialsOfAccount(ctx, accountID))
}

func (r *DefaultCredentialRepository) transform(res Credential) repository.Credential {
	return repository.Credential{
		AccountID: res.AccountID,
		Hash:      res.Hash,
		Type:      repository.CredentialType(res.Type),
		ExpiredAt: res.ExpiredAt,
	}
}

func (r *DefaultCredentialRepository) GetLast(
	ctx context.Context,
	accountID uuid.UUID,
	credType string,
) (repository.Credential, error) {
	res, err := r.queries.FindLastCredentialOfType(
		ctx, FindLastCredentialOfTypeParams{
			AccountID: accountID,
			Type:      credType,
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
