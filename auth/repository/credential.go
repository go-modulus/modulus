package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrCannotCreateCredential = errors.New("cannot create credential")
var ErrCredentialNotFound = errors.New("credential not found")

type Credential struct {
	IdentityID     uuid.UUID `json:"identityId"`
	CredentialHash string    `json:"credentialHash"`
	Type           string    `json:"type"`
	ExpiredAt      null.Time `json:"expiredAt"`
}

type CredentialRepository interface {
	// Create creates a new identity for the given user ID.
	// If the identity already exists, it returns github.com/go-modulus/modulus/auth/errors.ErrCredentialExists.
	// Otherwise, it returns nil.
	// The identity is a unique string that represents the user.
	// It is used for login and other operations.
	// It may be an email, username, or other unique identifier.
	// You are able to create multiple identities for a single user.
	Create(
		ctx context.Context,
		identityID uuid.UUID,
		credentialHash string,
		credType string,
		expiredAt *time.Time,
	) (Credential, error)

	// GetLast returns the last credential of the given type with the given identity ID.
	// If the credential does not exist, it returns github.com/go-modulus/modulus/auth.ErrCredentialNotFound.
	GetLast(
		ctx context.Context,
		identityID uuid.UUID,
		credType string,
	) (Credential, error)
}

type CredentialType string

const CredentialTypePassword CredentialType = "password"
const CredentialTypeOTP CredentialType = "otp"
