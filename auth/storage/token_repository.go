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

func (r *DefaultTokenRepository) RevokeAccountTokens(ctx context.Context, accountId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
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

	ident, err := r.queries.FindIdentityById(ctx, identityId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.AccessToken{}, errtrace.Wrap(repository.ErrIdentityNotFound)
		}
		return repository.AccessToken{}, errtrace.Wrap(err)
	}
	storedAccessToken, err := r.queries.CreateAccessToken(
		ctx,
		CreateAccessTokenParams{
			Hash:       accessToken,
			IdentityID: identityId,
			AccountID:  ident.AccountID,
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
	storedRefreshToken, err := r.queries.FindRefreshTokenByHash(ctx, refreshToken)
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
	storedAccessToken, err := r.queries.FindAccessTokenByHash(ctx, accessToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.AccessToken{}, errtrace.Wrap(repository.ErrTokenNotExist)
		}
		return repository.AccessToken{}, errtrace.Wrap(err)
	}
	return r.transformAccessToken(storedAccessToken), nil
}

//func (r *DefaultTokenRepository) RevokeAccessToken(ctx context.Context, accessToken string) error {
//	accessToken = r.hashToken(accessToken)
//	return r.queries.RevokeAccessToken(ctx, accessToken)
//}
//
//func (r *DefaultTokenRepository) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
//	refreshToken = r.hashToken(refreshToken)
//	return r.queries.RevokeRefreshToken(ctx, refreshToken)
//}

// ExpireTokens makes the valid tokens of the given session expired.
// It returns an error if the operation failed.
// Params:
//   - sessionId - the session where we want to expire the tokens.
//   - lag - the parameter is used to make the tokens expired in the future. It is helpful to avoid race conditions in the token refreshing.
//     We make tokens expired in several seconds to get the time for frontend to refresh the tokens in local storage and get the responses from simultaneous requests to the backend without error.
//   - tokenType - the type of tokens that we want to expire.
func (r *DefaultTokenRepository) ExpireTokens(
	ctx context.Context,
	sessionId uuid.UUID,
	lag time.Duration,
	tokenType repository.ExpirationTokenType,
) error {
	if tokenType == repository.AccessTokenType || tokenType == repository.BothTokenType {
		err := r.queries.ExpireSessionAccessTokens(
			ctx, ExpireSessionAccessTokensParams{
				ExpiresAt: time.Now().Add(lag),
				SessionID: sessionId,
			},
		)
		if err != nil {
			return errtrace.Wrap(err)
		}
	}
	if tokenType == repository.RefreshTokenType || tokenType == repository.BothTokenType {
		err := r.queries.ExpireSessionRefreshTokens(
			ctx, ExpireSessionRefreshTokensParams{
				ExpiresAt: time.Now().Add(lag),
				SessionID: sessionId,
			},
		)
		if err != nil {
			return errtrace.Wrap(err)
		}
	}

	return nil
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

func (r *DefaultTokenRepository) RevokeUserTokens(ctx context.Context, accountId uuid.UUID) error {
	sessionIds, err := r.queries.FindAccountNotRevokedSessionIds(ctx, accountId)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = r.queries.RevokeSessionsRefreshTokens(ctx, sessionIds)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = r.queries.RevokeAccountAccessTokens(ctx, accountId)
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
		AccountID:  storedAccessToken.AccountID,
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
