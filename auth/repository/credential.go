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
	Hash      string         `json:"hash"`
	AccountID uuid.UUID      `json:"accountId"`
	Type      CredentialType `json:"type"`
	ExpiredAt null.Time      `json:"expiredAt"`
}

type CredentialRepository interface {
	// Create creates a new credential for the given account ID.
	Create(
		ctx context.Context,
		accountID uuid.UUID,
		credentialHash string,
		credType CredentialType,
		expiredAt *time.Time,
	) (Credential, error)

	// GetLast returns the last credential of the given type with the given identity ID.
	// If the credential does not exist, it returns github.com/go-modulus/modulus/auth.ErrCredentialNotFound.
	GetLast(
		ctx context.Context,
		accountID uuid.UUID,
		credType string,
	) (Credential, error)

	// RemoveCredentials removes all credentials of the given account ID.
	RemoveCredentials(
		ctx context.Context,
		accountID uuid.UUID,
	) error
}

type CredentialType string

const CredentialTypePassword CredentialType = "password"
const CredentialTypeOTP CredentialType = "otp"
