package auth

import (
	"braces.dev/errtrace"
	"context"
	"encoding/json"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrIdentityExists = errors.New("identity exists")
var ErrIdentityNotFound = errors.New("identity not found")
var ErrCannotCreateIdentity = errors.New("cannot create identity")

type Identity struct {
	ID       uuid.UUID              `db:"id" json:"id"`
	Identity string                 `db:"identity" json:"identity"`
	UserID   uuid.UUID              `db:"user_id" json:"userId"`
	Roles    []string               `db:"roles" json:"roles"`
	Status   IdentityStatus         `db:"status" json:"status"`
	Data     map[string]interface{} `db:"data" json:"data"`
}

func (i Identity) IsBlocked() bool {
	return i.Status == IdentityStatusBlocked
}

type IdentityStatus string

const (
	IdentityStatusActive  IdentityStatus = "active"
	IdentityStatusBlocked IdentityStatus = "blocked"
)

type IdentityRepository interface {
	// Create creates a new identity for the given user ID.
	// If the identity already exists, it returns github.com/go-modulus/modulus/auth.ErrIdentityExists.
	// If the identity cannot be created, it returns github.com/go-modulus/modulus/auth.ErrCannotCreateIdentity.
	//
	// The identity is a unique string that represents the user.
	// It is used for login and other operations.
	// It may be an email, username, or other unique identifier.
	// You are able to create multiple identities for a single user.
	Create(
		ctx context.Context,
		identity string,
		userId uuid.UUID,
		AdditionalData map[string]interface{},
	) (Identity, error)

	// Get returns the identity with the given identity string.
	// If the identity does not exist, it returns github.com/go-modulus/modulus/auth.ErrIdentityNotFound.
	Get(
		ctx context.Context,
		identity string,
	) (Identity, error)
	// GetById returns the identity with the given ID.
	// If the identity does not exist, it returns github.com/go-modulus/modulus/auth.ErrIdentityNotFound.
	GetById(
		ctx context.Context,
		id uuid.UUID,
	) (Identity, error)

	AddRoles(
		ctx context.Context,
		identityID uuid.UUID,
		roles ...string,
	) error

	RemoveRoles(
		ctx context.Context,
		identityID uuid.UUID,
		roles ...string,
	) error
}

type DefaultIdentityRepository struct {
	queries *storage.Queries
}

func NewDefaultIdentityRepository(db *pgxpool.Pool) IdentityRepository {
	return &DefaultIdentityRepository{
		queries: storage.New(db),
	}
}

func (r *DefaultIdentityRepository) Create(
	ctx context.Context,
	identity string,
	UserID uuid.UUID,
	AdditionalData map[string]interface{},
) (Identity, error) {
	_, err := r.Get(ctx, identity)
	if err == nil {
		return Identity{}, ErrIdentityExists
	} else if !errors.Is(err, ErrIdentityNotFound) {
		return Identity{}, errtrace.Wrap(err)
	}
	id := uuid.Must(uuid.NewV6())
	if AdditionalData == nil {
		AdditionalData = make(map[string]interface{})
	}
	dataVal, err := json.Marshal(AdditionalData)
	if err != nil {
		return Identity{}, errtrace.Wrap(err)
	}
	storedIdentity, err := r.queries.CreateIdentity(
		ctx, storage.CreateIdentityParams{
			ID:       id,
			Identity: identity,
			UserID:   UserID,
			Data:     dataVal,
		},
	)

	if err != nil {
		return Identity{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateIdentity, err))
	}

	return r.transform(storedIdentity), nil
}

func (r *DefaultIdentityRepository) transform(
	identity storage.Identity,
) Identity {
	var data map[string]interface{}
	if err := json.Unmarshal(identity.Data, &data); err != nil {
		data = make(map[string]interface{})
	}
	return Identity{
		ID:       identity.ID,
		Identity: identity.Identity,
		UserID:   identity.UserID,
		Roles:    identity.Roles,
		Status:   IdentityStatus(identity.Status),
		Data:     data,
	}
}

func (r *DefaultIdentityRepository) Get(
	ctx context.Context,
	identity string,
) (Identity, error) {
	res, err := r.queries.FindIdentity(ctx, identity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, ErrIdentityNotFound
		}
		return Identity{}, errtrace.Wrap(err)
	}
	return r.transform(res), nil
}

func (r *DefaultIdentityRepository) GetById(
	ctx context.Context,
	id uuid.UUID,
) (Identity, error) {
	res, err := r.queries.FindIdentityById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, ErrIdentityNotFound
		}
		return Identity{}, errtrace.Wrap(err)
	}
	return r.transform(res), nil
}

func (r *DefaultIdentityRepository) AddRoles(
	ctx context.Context,
	identityID uuid.UUID,
	roles ...string,
) error {
	return r.queries.AddRoles(
		ctx, storage.AddRolesParams{
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
		ctx, storage.RemoveRolesParams{
			ID:    identityID,
			Roles: roles,
		},
	)
}
