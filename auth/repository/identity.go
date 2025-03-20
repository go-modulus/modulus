package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
)

type IdentityType string

const (
	IdentityTypeEmail    IdentityType = "email"
	IdentityTypePhone    IdentityType = "phone"
	IdentityTypeNickname IdentityType = "nickname"
)

var ErrIdentityExists = errors.New("identity exists")
var ErrIdentityNotFound = errors.New("identity not found")
var ErrCannotCreateIdentity = errors.New("cannot create identity")

type Identity struct {
	ID        uuid.UUID              `db:"id" json:"id"`
	Identity  string                 `db:"identity" json:"identity"`
	AccountID uuid.UUID              `db:"user_id" json:"accountId"`
	Type      IdentityType           `db:"type" json:"type"`
	Status    IdentityStatus         `db:"status" json:"status"`
	Data      map[string]interface{} `db:"data" json:"data"`
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
	// Create creates a new identity for the given account ID.
	// If the identity already exists, it returns github.com/go-modulus/modulus/auth.ErrIdentityExists.
	// If the identity cannot be created, it returns github.com/go-modulus/modulus/auth.ErrCannotCreateIdentity.
	//
	// The identity is a unique string that represents the user.
	// It is used for login and other operations.
	// It may be an email, username, or other unique identifier.
	// You are able to create multiple identities for a single account.
	Create(
		ctx context.Context,
		identity string,
		accountID uuid.UUID,
		identityType IdentityType,
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

	// GetByAccountID returns the identities with the given account ID.
	GetByAccountID(
		ctx context.Context,
		accountID uuid.UUID,
	) ([]Identity, error)

	// RemoveAccountIdentities removes the identities with the given account ID.
	RemoveAccountIdentities(
		ctx context.Context,
		accountID uuid.UUID,
	) error

	// RemoveIdentity removes the identity.
	RemoveIdentity(
		ctx context.Context,
		identity string,
	) error

	// BlockIdentity blocks the identity.
	BlockIdentity(
		ctx context.Context,
		identity string,
	) error
}
