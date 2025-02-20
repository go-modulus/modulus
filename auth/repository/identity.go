package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
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
		userID uuid.UUID,
		additionalData map[string]interface{},
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
