package storage

import (
	"braces.dev/errtrace"
	"context"
	"encoding/json"
	"github.com/go-modulus/modulus/auth/hash"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type DefaultTokenRepository struct {
	queries      *Queries
	hashStrategy hash.TokenHashStrategy
}

func NewDefaultTokenRepository(
	db *pgxpool.Pool,
	hashStrategy hash.TokenHashStrategy,
) repository.TokenRepository {
	return &DefaultTokenRepository{
		queries:      New(db),
		hashStrategy: hashStrategy,
	}
}

func (r *DefaultTokenRepository) CreateAccessToken(
	ctx context.Context,
	accessToken string,
	identityId uuid.UUID,
	userId uuid.UUID,
	roles []string,
	sessionId uuid.UUID,
	data map[string]interface{},
	expiresAt time.Time,
) (repository.AccessToken, error) {
	accessToken = r.hashToken(accessToken)

	if data == nil {
		data = make(map[string]interface{})
	}
	dataJson, err := json.Marshal(data)
	if err != nil {
		return repository.AccessToken{}, errtrace.Wrap(errors.WithCause(repository.ErrCannotCreateAccessToken, err))
	}

	storedAccessToken, err := r.queries.CreateAccessToken(
		ctx,
		CreateAccessTokenParams{
			Hash:       accessToken,
			IdentityID: identityId,
			UserID:     userId,
			Roles:      roles,
			SessionID:  sessionId,
			Data:       dataJson,
			ExpiresAt:  expiresAt,
		},
	)
	if err != nil {
		return repository.AccessToken{}, err
	}
	return r.transformAccessToken(storedAccessToken), nil
}

func (r *DefaultTokenRepository) CreateRefreshToken(
	ctx context.Context,
	refreshToken string,
	sessionId uuid.UUID,
	identityID uuid.UUID,
	expiresAt time.Time,
) (repository.RefreshToken, error) {
	refreshToken = r.hashToken(refreshToken)
	storedRefreshToken, err := r.queries.CreateRefreshToken(
		ctx,
		CreateRefreshTokenParams{
			Hash:       refreshToken,
			SessionID:  sessionId,
			ExpiresAt:  expiresAt,
			IdentityID: identityID,
		},
	)
	if err != nil {
		return repository.RefreshToken{}, errtrace.Wrap(errors.WithCause(repository.ErrCannotCreateRefreshToken, err))
	}
	return r.transformRefreshToken(storedRefreshToken), nil
}

func (r *DefaultTokenRepository) GetRefreshToken(ctx context.Context, refreshToken string) (
	repository.RefreshToken,
	error,
) {
	refreshToken = r.hashToken(refreshToken)
	storedRefreshToken, err := r.queries.GetRefreshTokenByHash(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.RefreshToken{}, errtrace.Wrap(repository.ErrTokenNotExist)
		}
		return repository.RefreshToken{}, errtrace.Wrap(err)
	}
	return r.transformRefreshToken(storedRefreshToken), nil
}

func (r *DefaultTokenRepository) GetAccessToken(ctx context.Context, accessToken string) (
	repository.AccessToken,
	error,
) {
	accessToken = r.hashToken(accessToken)
	storedAccessToken, err := r.queries.GetAccessTokenByHash(ctx, accessToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.AccessToken{}, errtrace.Wrap(repository.ErrTokenNotExist)
		}
		return repository.AccessToken{}, errtrace.Wrap(err)
	}
	return r.transformAccessToken(storedAccessToken), nil
}

func (r *DefaultTokenRepository) RevokeAccessToken(ctx context.Context, accessToken string) error {
	accessToken = r.hashToken(accessToken)
	return r.queries.RevokeAccessToken(ctx, accessToken)
}

func (r *DefaultTokenRepository) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	refreshToken = r.hashToken(refreshToken)
	return r.queries.RevokeRefreshToken(ctx, refreshToken)
}

func (r *DefaultTokenRepository) RevokeSessionTokens(ctx context.Context, sessionId uuid.UUID) error {
	err := r.queries.RevokeSessionAccessTokens(ctx, sessionId)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = r.queries.RevokeSessionRefreshTokens(ctx, sessionId)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *DefaultTokenRepository) RevokeUserTokens(ctx context.Context, userId uuid.UUID) error {
	sessionIds, err := r.queries.GetUserNotRevokedSessionIds(ctx, userId)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = r.queries.RevokeSessionsRefreshTokens(ctx, sessionIds)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = r.queries.RevokeUserAccessTokens(ctx, userId)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return nil
}

func (r *DefaultTokenRepository) transformAccessToken(storedAccessToken AccessToken) repository.AccessToken {
	var data map[string]interface{}
	if err := json.Unmarshal(storedAccessToken.Data, &data); err != nil {
		data = make(map[string]interface{})
	}
	return repository.AccessToken{
		Hash:       storedAccessToken.Hash,
		IdentityID: storedAccessToken.IdentityID,
		UserID:     storedAccessToken.UserID,
		Roles:      storedAccessToken.Roles,
		SessionID:  storedAccessToken.SessionID,
		Data:       data,
		RevokedAt:  storedAccessToken.RevokedAt,
		ExpiresAt:  storedAccessToken.ExpiresAt,
	}
}

func (r *DefaultTokenRepository) transformRefreshToken(storedRefreshToken RefreshToken) repository.RefreshToken {
	return repository.RefreshToken{
		Hash:       storedRefreshToken.Hash,
		IdentityID: storedRefreshToken.IdentityID,
		SessionID:  storedRefreshToken.SessionID,
		RevokedAt:  storedRefreshToken.RevokedAt,
		ExpiresAt:  storedRefreshToken.ExpiresAt,
	}
}

func (r *DefaultTokenRepository) hashToken(token string) string {
	return r.hashStrategy.Hash(token)
}
