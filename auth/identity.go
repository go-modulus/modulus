package auth

import (
	"context"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/gofrs/uuid"
)

type IdentityRepository interface {
	// MakeIdentity creates a new identity for the given user ID.
	// If the identity already exists, it returns github.com/go-modulus/modulus/auth/errors.ErrIdentityExists.
	// Otherwise, it returns nil.
	// The identity is a unique string that represents the user.
	// It is used for login and other operations.
	// It may be an email, username, or other unique identifier.
	// You are able to create multiple identities for a single user.
	MakeIdentity(ctx context.Context, identity string, userId uuid.UUID, AdditionalData map[string]interface{}) error

	GetIdentity(
		ctx context.Context,
		identity string,
	) (storage.Identity, error)
}

func NewIdentityRepository(defRepo *storage.DefaultRepository) IdentityRepository {
	return defRepo
}
