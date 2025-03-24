package storage

import (
	"braces.dev/errtrace"
	"context"
	"encoding/json"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DefaultIdentityRepository struct {
	queries *Queries
}

func NewDefaultIdentityRepository(db *pgxpool.Pool) repository.IdentityRepository {
	return &DefaultIdentityRepository{
		queries: New(db),
	}
}

func (r *DefaultIdentityRepository) Create(
	ctx context.Context,
	identity string,
	accountID uuid.UUID,
	identityType repository.IdentityType,
	additionalData map[string]interface{},
) (repository.Identity, error) {
	_, err := r.Get(ctx, identity)
	if err == nil {
		return repository.Identity{}, repository.ErrIdentityExists
	} else if !errors.Is(err, repository.ErrIdentityNotFound) {
		return repository.Identity{}, errtrace.Wrap(err)
	}
	id := uuid.Must(uuid.NewV6())
	var dataVal []byte
	if len(additionalData) > 0 {
		dataVal, err = json.Marshal(additionalData)
		if err != nil {
			return repository.Identity{}, errtrace.Wrap(err)
		}
	}

	storedIdentity, err := r.queries.CreateIdentity(
		ctx, CreateIdentityParams{
			ID:        id,
			Identity:  identity,
			AccountID: accountID,
			Data:      dataVal,
			Type:      string(identityType),
		},
	)

	if err != nil {
		return repository.Identity{}, errtrace.Wrap(errors.WithCause(repository.ErrCannotCreateIdentity, err))
	}

	return r.Transform(storedIdentity), nil
}

func (r *DefaultIdentityRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) (
	[]repository.Identity,
	error,
) {
	idents, err := r.queries.FindAccountIdentities(ctx, accountID)
	if err != nil {
		return nil, errtrace.Wrap(err)
	}
	var res []repository.Identity
	for _, ident := range idents {
		res = append(res, r.Transform(ident))
	}
	return res, nil
}

func (r *DefaultIdentityRepository) RemoveAccountIdentities(ctx context.Context, accountID uuid.UUID) error {
	return errtrace.Wrap(r.queries.RemoveIdentitiesOfAccount(ctx, accountID))
}

func (r *DefaultIdentityRepository) RemoveIdentity(ctx context.Context, identity string) error {
	ident, err := r.Get(ctx, identity)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return errtrace.Wrap(r.queries.RemoveIdentity(ctx, ident.ID))
}

func (r *DefaultIdentityRepository) BlockIdentity(ctx context.Context, identity string) error {
	ident, err := r.Get(ctx, identity)
	if err != nil {
		return errtrace.Wrap(err)
	}

	return errtrace.Wrap(r.queries.BlockIdentity(ctx, ident.ID))
}

func (r *DefaultIdentityRepository) Transform(
	identity Identity,
) repository.Identity {
	var data map[string]interface{}
	if err := json.Unmarshal(identity.Data, &data); err != nil {
		data = make(map[string]interface{})
	}
	return repository.Identity{
		ID:        identity.ID,
		Identity:  identity.Identity,
		AccountID: identity.AccountID,
		Status:    repository.IdentityStatus(identity.Status),
		Data:      data,
		Type:      repository.IdentityType(identity.Type),
	}
}

func (r *DefaultIdentityRepository) Get(
	ctx context.Context,
	identity string,
) (repository.Identity, error) {
	res, err := r.queries.FindIdentity(ctx, identity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Identity{}, repository.ErrIdentityNotFound
		}
		return repository.Identity{}, errtrace.Wrap(err)
	}
	return r.Transform(res), nil
}

func (r *DefaultIdentityRepository) GetById(
	ctx context.Context,
	id uuid.UUID,
) (repository.Identity, error) {
	res, err := r.queries.FindIdentityById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Identity{}, repository.ErrIdentityNotFound
		}
		return repository.Identity{}, errtrace.Wrap(err)
	}
	return r.Transform(res), nil
}
