package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
)

var ErrAccountExists = errors.New("account exists")
var ErrAccountNotFound = errors.New("account not found")
var ErrCannotCreateAccount = errors.New("cannot create account")

type Account struct {
	ID     uuid.UUID     `db:"id" json:"id"`
	Roles  []string      `db:"roles" json:"roles"`
	Status AccountStatus `db:"status" json:"status"`
}

func (i Account) IsBlocked() bool {
	return i.Status == AccountStatusBlocked
}

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "active"
	AccountStatusBlocked AccountStatus = "blocked"
)

type AccountRepository interface {
	// Create creates a single new authorization account for the user.
	Create(
		ctx context.Context,
		ID uuid.UUID,
	) (Account, error)

	// Get returns the account by its ID.
	// If the identity does not exist, it returns github.com/go-modulus/modulus/auth.ErrAccountNotFound.
	Get(
		ctx context.Context,
		ID uuid.UUID,
	) (Account, error)

	AddRoles(
		ctx context.Context,
		ID uuid.UUID,
		roles ...string,
	) error

	RemoveRoles(
		ctx context.Context,
		ID uuid.UUID,
		roles ...string,
	) error

	// RemoveAccount removes the identity.
	RemoveAccount(
		ctx context.Context,
		ID uuid.UUID,
	) error

	// BlockAccount blocks the identity.
	BlockAccount(
		ctx context.Context,
		ID uuid.UUID,
	) error
}
