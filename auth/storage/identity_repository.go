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
	userID uuid.UUID,
	additionalData map[string]interface{},
) (repository.Identity, error) {
	_, err := r.Get(ctx, identity)
	if err == nil {
		return repository.Identity{}, repository.ErrIdentityExists
	} else if !errors.Is(err, repository.ErrIdentityNotFound) {
		return repository.Identity{}, errtrace.Wrap(err)
	}
	id := uuid.Must(uuid.NewV6())
	if additionalData == nil {
		additionalData = make(map[string]interface{})
	}
	dataVal, err := json.Marshal(additionalData)
	if err != nil {
		return repository.Identity{}, errtrace.Wrap(err)
	}
	storedIdentity, err := r.queries.CreateIdentity(
		ctx, CreateIdentityParams{
			ID:       id,
			Identity: identity,
			UserID:   userID,
			Data:     dataVal,
		},
	)

	if err != nil {
		return repository.Identity{}, errtrace.Wrap(errors.WrapCause(repository.ErrCannotCreateIdentity, err))
	}

	return r.Transform(storedIdentity), nil
}

func (r *DefaultIdentityRepository) Transform(
	identity Identity,
) repository.Identity {
	var data map[string]interface{}
	if err := json.Unmarshal(identity.Data, &data); err != nil {
		data = make(map[string]interface{})
	}
	return repository.Identity{
		ID:       identity.ID,
		Identity: identity.Identity,
		UserID:   identity.UserID,
		Roles:    identity.Roles,
		Status:   repository.IdentityStatus(identity.Status),
		Data:     data,
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

func (r *DefaultIdentityRepository) AddRoles(
	ctx context.Context,
	identityID uuid.UUID,
	roles ...string,
) error {
	return r.queries.AddRoles(
		ctx, AddRolesParams{
			ID:    identityID,
			Roles: roles,
		},
	)
}

func (r *DefaultIdentityRepository) RemoveRoles(
	ctx context.Context,
	identityID uuid.UUID,
	roles ...string,
) error {
	return r.queries.RemoveRoles(
		ctx, RemoveRolesParams{
			ID:    identityID,
			Roles: roles,
		},
	)
}
