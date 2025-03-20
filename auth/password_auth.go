package auth

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidIdentity = errors.New("invalid identity")
var ErrIdentityIsBlocked = errors.New("identity is blocked")
var ErrInvalidPassword = errors.New("invalid password")
var ErrCannotHashPassword = errors.New("cannot hash password")

type PasswordAuthenticator struct {
	accountRepository    repository.AccountRepository
	identityRepository   repository.IdentityRepository
	credentialRepository repository.CredentialRepository
}

func NewPasswordAuthenticator(
	accountRepository repository.AccountRepository,
	identityRepository repository.IdentityRepository,
	credentialRepository repository.CredentialRepository,
) *PasswordAuthenticator {
	return &PasswordAuthenticator{
		accountRepository:    accountRepository,
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
// * github.com/go-modulus/modulus/auth.ErrInvalidIdentity - if identity is not found in the repository.
func (a *PasswordAuthenticator) Authenticate(ctx context.Context, identity, password string) (Performer, error) {
	identityObj, err := a.identityRepository.Get(ctx, identity)
	if err != nil {
		if errors.Is(err, repository.ErrIdentityNotFound) {
			return Performer{}, errtrace.Wrap(ErrInvalidIdentity)
		}
		return Performer{}, errtrace.Wrap(err)
	}

	if identityObj.IsBlocked() {
		return Performer{}, errtrace.Wrap(ErrIdentityIsBlocked)
	}

	cred, err := a.credentialRepository.GetLast(ctx, identityObj.AccountID, string(repository.CredentialTypePassword))
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return Performer{}, errtrace.Wrap(ErrInvalidPassword)
		}
		return Performer{}, errtrace.Wrap(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(cred.Hash), []byte(password))
	if err != nil {
		return Performer{}, errtrace.Wrap(ErrInvalidPassword)
	}
	return Performer{ID: identityObj.AccountID, SessionID: uuid.Must(uuid.NewV6()), IdentityID: identityObj.ID}, nil
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
	identityType repository.IdentityType,
	roles []string,
	additionalData map[string]interface{},
) (repository.Account, error) {
	identityObj, err := a.identityRepository.Get(ctx, identity)
	if err == nil {
		if identityObj.IsBlocked() {
			return repository.Account{}, errtrace.Wrap(ErrIdentityIsBlocked)
		}
		return repository.Account{}, errtrace.Wrap(repository.ErrIdentityExists)
	} else if !errors.Is(err, repository.ErrIdentityNotFound) {
		return repository.Account{}, errtrace.Wrap(err)
	}

	accountID, err := uuid.NewV6()
	if err != nil {
		return repository.Account{}, errtrace.Wrap(err)
	}

	account, err := a.accountRepository.Create(ctx, accountID)
	if err != nil {
		return repository.Account{}, errtrace.Wrap(err)
	}

	identityObj, err = a.identityRepository.Create(ctx, identity, accountID, identityType, additionalData)
	if err != nil {
		return repository.Account{}, errtrace.Wrap(err)
	}

	defer func() {
		if err != nil {
			_ = a.accountRepository.RemoveAccount(ctx, accountID)
			_ = a.identityRepository.RemoveIdentity(ctx, identity)
		}
	}()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return repository.Account{}, errtrace.Wrap(errors.WithCause(ErrCannotHashPassword, err))
	}

	_, err = a.credentialRepository.Create(
		ctx,
		accountID,
		string(hash),
		repository.CredentialTypePassword,
		nil,
	)
	if err != nil {
		return repository.Account{}, errtrace.Wrap(err)
	}

	if len(roles) > 0 {
		err = a.accountRepository.AddRoles(ctx, identityObj.ID, roles...)
		if err != nil {
			return repository.Account{}, errtrace.Wrap(err)
		}
	}

	return account, nil
}

func (a *PasswordAuthenticator) RemoveIdentity(ctx context.Context, identity string) error {
	ident, err := a.identityRepository.Get(ctx, identity)
	if err != nil {
		if errors.Is(err, repository.ErrIdentityNotFound) {
			return nil
		}
	}

	err = a.identityRepository.RemoveIdentity(ctx, identity)
	if err != nil {
		return errtrace.Wrap(err)
	}

	identities, err := a.identityRepository.GetByAccountID(ctx, ident.AccountID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	if len(identities) == 0 {
		err = a.accountRepository.RemoveAccount(ctx, ident.AccountID)
		if err != nil {
			return errtrace.Wrap(err)
		}
	}

	return nil
}
