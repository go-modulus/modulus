package auth

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrIdentityIsBlocked = errors.New("identity is blocked")
var ErrInvalidPassword = errors.New("invalid password")
var ErrCannotHashPassword = errors.New("cannot hash password")

type PasswordAuthenticator struct {
	identityRepository   IdentityRepository
	credentialRepository CredentialRepository
}

func NewPasswordAuthenticator(
	identityRepository IdentityRepository,
	credentialRepository CredentialRepository,
) *PasswordAuthenticator {
	return &PasswordAuthenticator{
		identityRepository:   identityRepository,
		credentialRepository: credentialRepository,
	}
}

// Authenticate authenticates the user with the given identity and password.
// It returns the performer of the authenticated user.
//
// Errors:
// * github.com/go-modulus/modulus/auth.ErrIdentityIsBlocked - if the identity is blocked.
// * github.com/go-modulus/modulus/auth.ErrInvalidPassword - if the password is invalid.
// * Any error from the IdentityRepository.Get method (e.g. github.com/go-modulus/modulus/auth.ErrIdentityNotFound).
// * Any error from the CredentialRepository.GetLast method (e.g. github.com/go-modulus/modulus/auth.ErrCredentialNotFound).
func (a *PasswordAuthenticator) Authenticate(ctx context.Context, identity, password string) (Performer, error) {
	identityObj, err := a.identityRepository.Get(ctx, identity)
	if err != nil {
		return Performer{}, errtrace.Wrap(err)
	}

	if identityObj.IsBlocked() {
		return Performer{}, errtrace.Wrap(ErrIdentityIsBlocked)
	}

	cred, err := a.credentialRepository.GetLast(ctx, identityObj.ID, string(CredentialTypePassword))
	if err != nil {
		return Performer{}, errtrace.Wrap(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(cred.CredentialHash), []byte(password))
	if err != nil {
		return Performer{}, errtrace.Wrap(ErrInvalidPassword)
	}
	return Performer{ID: identityObj.UserID, SessionID: uuid.Must(uuid.NewV6())}, nil
}

// Register registers a new user account with the given identity and password.
// In the additionalData, you can pass any additional data you want to store (e.g. IP, unregistered user token from frontend, etc.).
// It returns the performer of the registered user.
//
// Errors:
// * github.com/go-modulus/modulus/auth.ErrIdentityExists - if you try to register a user for the same identity.
// * github.com/go-modulus/modulus/auth.ErrIdentityIsBlocked - if the identity exists in the storage, and it has status blocked.
// * Any error from the IdentityRepository.Create method.
// * Any error from the CredentialRepository.Create method.
func (a *PasswordAuthenticator) Register(
	ctx context.Context,
	identity,
	password string,
	userID uuid.UUID,
	roles []string,
	additionalData map[string]interface{},
) (Identity, error) {
	identityObj, err := a.identityRepository.Get(ctx, identity)
	if err == nil {
		if identityObj.IsBlocked() {
			return Identity{}, errtrace.Wrap(ErrIdentityIsBlocked)
		}
		return Identity{}, errtrace.Wrap(ErrIdentityExists)
	} else if !errors.Is(err, ErrIdentityNotFound) {
		return Identity{}, errtrace.Wrap(err)
	}

	identityObj, err = a.identityRepository.Create(ctx, identity, userID, additionalData)
	if err != nil {
		return Identity{}, errtrace.Wrap(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Identity{}, errtrace.Wrap(errors.WrapCause(ErrCannotHashPassword, err))
	}

	_, err = a.credentialRepository.Create(ctx, identityObj.ID, string(hash), string(CredentialTypePassword), nil)
	if err != nil {
		return Identity{}, errtrace.Wrap(err)
	}

	if len(roles) > 0 {
		err = a.identityRepository.AddRoles(ctx, identityObj.ID, roles...)
		if err != nil {
			return Identity{}, errtrace.Wrap(err)
		}
	}

	return identityObj, nil
}
