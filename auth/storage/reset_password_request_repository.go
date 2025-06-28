package storage

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
	"time"
)

type ResetPasswordConfig struct {
	ResetPasswordLife time.Duration `env:"AUTH_RESET_PASSWORD_LIFE, default=1h"`
	ResendCooldown    time.Duration `env:"AUTH_RESET_PASSWORD_RESEND_COOLDOWN, default=5m"`
}

type DefaultResetPasswordRequestRepository struct {
	queries *Queries
	config  ResetPasswordConfig
}

func NewDefaultResetPasswordRequestRepository(
	db *pgxpool.Pool,
	config ResetPasswordConfig,
) repository.ResetPasswordRequestRepository {
	return &DefaultResetPasswordRequestRepository{
		queries: New(db),
		config:  config,
	}
}

func (r *DefaultResetPasswordRequestRepository) GetActiveRequest(
	ctx context.Context,
	accountID uuid.UUID,
) (repository.ResetPasswordRequest, error) {
	res, err := r.queries.GetReadyForUseResetPasswordRequest(ctx, accountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ResetPasswordRequest{}, repository.ErrResetPasswordRequestNotFound
		}
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}
	return r.Transform(res), nil
}

func (r *DefaultResetPasswordRequestRepository) ExpireRequest(ctx context.Context, ID uuid.UUID) error {
	err := r.queries.ExpireResetPasswordRequest(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *DefaultResetPasswordRequestRepository) CreateResetPassword(
	ctx context.Context,
	id uuid.UUID,
	accountID uuid.UUID,
	token string,
) (repository.ResetPasswordRequest, error) {
	res, err := r.queries.CreateResetPasswordRequest(
		ctx, CreateResetPasswordRequestParams{
			ID:        id,
			AccountID: accountID,
			Token:     token,
		},
	)
	if err != nil {
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}
	return r.Transform(res), nil
}

func (r *DefaultResetPasswordRequestRepository) UpdateLastSent(ctx context.Context, ID uuid.UUID) error {
	err := r.queries.UpdateLastSentRequest(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *DefaultResetPasswordRequestRepository) GetResetPasswordByToken(
	ctx context.Context,
	token string,
) (repository.ResetPasswordRequest, error) {
	res, err := r.queries.GetResetPasswordRequestByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.ResetPasswordRequest{}, repository.ErrResetPasswordRequestNotFound
		}
		return repository.ResetPasswordRequest{}, errtrace.Wrap(err)
	}
	return r.Transform(res), nil
}

func (r *DefaultResetPasswordRequestRepository) UseResetPassword(ctx context.Context, ID uuid.UUID) error {
	err := r.queries.UseResetPasswordRequest(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *DefaultResetPasswordRequestRepository) Transform(
	request ResetPasswordRequest,
) repository.ResetPasswordRequest {
	isUsed := request.UsedAt.Valid
	aliveTill := request.CreatedAt.Add(r.config.ResetPasswordLife)
	coolDownTill := null.Time{}
	if request.LastSendAt.Valid {
		coolDownTill = null.TimeFrom(request.LastSendAt.Time.Add(r.config.ResendCooldown))
	}
	return repository.ResetPasswordRequest{
		ID:           request.ID,
		AccountID:    request.AccountID,
		Status:       repository.ResetPasswordStatus(request.Status),
		Token:        request.Token,
		IsUsed:       isUsed,
		AliveTill:    aliveTill,
		CoolDownTill: coolDownTill,
	}
}
