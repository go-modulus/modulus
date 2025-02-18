package auth

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrCannotCreateCredential = errors.New("cannot create credential")
var ErrCredentialNotFound = errors.New("credential not found")

type Credential struct {
	ID             uuid.UUID `db:"id" json:"id"`
	IdentityID     uuid.UUID `db:"identity_id" json:"identityId"`
	CredentialHash string    `db:"credential_hash" json:"credentialHash"`
	Type           string    `db:"type" json:"type"`
	ExpiredAt      null.Time `db:"expired_at" json:"expiredAt"`
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

type DefaultCredentialRepository struct {
	queries *storage.Queries
}

func NewDefaultCredentialRepository(db *pgxpool.Pool) CredentialRepository {
	return &DefaultCredentialRepository{
		queries: storage.New(db),
	}
}

func (r *DefaultCredentialRepository) Create(
	ctx context.Context,
	identityID uuid.UUID,
	credHash string,
	credType string,
	expiredAt *time.Time,
) (Credential, error) {
	expAt := null.TimeFromPtr(expiredAt)
	cred, err := r.queries.CreateCredential(
		ctx, storage.CreateCredentialParams{
			ID:             uuid.Must(uuid.NewV6()),
			IdentityID:     identityID,
			Type:           credType,
			CredentialHash: credHash,
			ExpiredAt:      expAt,
		},
	)

	if err != nil {
		return Credential{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateCredential, err))
	}

	return r.transform(cred), nil
}

func (r *DefaultCredentialRepository) transform(res storage.Credential) Credential {
	return Credential{
		ID:             res.ID,
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
) (Credential, error) {
	res, err := r.queries.FindLastCredentialOfType(
		ctx, storage.FindLastCredentialOfTypeParams{
			IdentityID: identityID,
			Type:       credType,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Credential{}, ErrCredentialNotFound
		}
		return Credential{}, errtrace.Wrap(err)
	}
	return r.transform(res), nil
}
