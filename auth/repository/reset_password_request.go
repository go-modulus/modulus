package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrResetPasswordRequestNotFound = errsys.New(
	"reset password request not found",
	"A request for resetting password was not found",
)

type ResetPasswordStatus string

const (
	ResetPasswordStatusActive  ResetPasswordStatus = "active"
	ResetPasswordStatusExpired ResetPasswordStatus = "expired"
	ResetPasswordStatusUsed    ResetPasswordStatus = "used"
)

type ResetPasswordRequest struct {
	ID           uuid.UUID           `json:"id"`
	AccountID    uuid.UUID           `json:"accountId"`
	Status       ResetPasswordStatus `json:"status"`
	Token        string              `json:"token"`
	IsUsed       bool                `json:"isUsed"`
	AliveTill    time.Time           `json:"lifePeriod"`
	CoolDownTill null.Time           `json:"coolDownPeriod"`
}

func (r ResetPasswordRequest) IsAlive() bool {
	return r.AliveTill.After(time.Now())
}

func (r ResetPasswordRequest) CanBeResent() bool {
	return !r.CoolDownTill.Valid || r.CoolDownTill.Time.Before(time.Now())
}

type ResetPasswordRequestRepository interface {
	// GetActiveRequest retrieves an active reset password request for the given account ID.
	// If no such request exists, it returns ErrResetPasswordRequestNotFound.
	GetActiveRequest(ctx context.Context, accountID uuid.UUID) (ResetPasswordRequest, error)
	ExpireRequest(ctx context.Context, ID uuid.UUID) error
	CreateResetPassword(ctx context.Context, id uuid.UUID, accountID uuid.UUID, token string) (
		ResetPasswordRequest,
		error,
	)
	UpdateLastSent(ctx context.Context, ID uuid.UUID) error
	// GetResetPasswordByToken retrieves a reset password request by its token.
	// If no such request exists, it returns ErrResetPasswordRequestNotFound.
	GetResetPasswordByToken(ctx context.Context, token string) (ResetPasswordRequest, error)
	UseResetPassword(ctx context.Context, ID uuid.UUID) error
}
