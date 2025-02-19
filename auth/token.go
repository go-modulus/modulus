package auth

import (
	"braces.dev/errtrace"
	"context"
	"crypto/sha1"
	"encoding/json"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
	"time"
)

type AccessToken struct {
	Hash       string                 `json:"hash"`
	IdentityID uuid.UUID              `json:"identityId"`
	SessionID  uuid.UUID              `json:"sessionId"`
	UserID     uuid.UUID              `json:"userId"`
	Roles      []string               `json:"roles"`
	Data       map[string]interface{} `json:"data"`
	RevokedAt  null.Time              `json:"revokedAt"`
	ExpiresAt  time.Time              `json:"expiresAt"`
}

type RefreshToken struct {
	Hash      string    `json:"hash"`
	SessionID uuid.UUID `json:"sessionId"`
	RevokedAt null.Time `json:"revokedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type TokenRepository interface {
	CreateAccessToken(
		ctx context.Context,
		accessToken string,
		identityId uuid.UUID,
		userId uuid.UUID,
		roles []string,
		sessionId uuid.UUID,
		data map[string]interface{},
		expiresAt time.Time,
	) (AccessToken, error)
	CreateRefreshToken(
		ctx context.Context,
		refreshToken string,
		sessionId uuid.UUID,
		expiresAt time.Time,
	) (RefreshToken, error)
	GetRefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error)
	GetAccessToken(ctx context.Context, accessToken string) (AccessToken, error)
	RevokeAccessToken(ctx context.Context, accessToken string) error
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	RevokeSessionTokens(ctx context.Context, sessionId uuid.UUID) error
	RevokeUserTokens(ctx context.Context, userId uuid.UUID) error
}

type DefaultTokenRepository struct {
	queries *storage.Queries
	config  ModuleConfig
}

func NewDefaultTokenRepository(
	db *pgxpool.Pool,
	config ModuleConfig,
) TokenRepository {
	return &DefaultTokenRepository{
		queries: storage.New(db),
		config:  config,
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
) (AccessToken, error) {
	accessToken = r.hashToken(accessToken)

	dataJson, err := json.Marshal(data)
	if err != nil {
		return AccessToken{}, errtrace.Wrap(err)
	}

	storedAccessToken, err := r.queries.CreateAccessToken(
		ctx,
		storage.CreateAccessTokenParams{
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
		return AccessToken{}, err
	}
	return r.transformAccessToken(storedAccessToken), nil
}

func (r *DefaultTokenRepository) CreateRefreshToken(
	ctx context.Context,
	refreshToken string,
	sessionId uuid.UUID,
	expiresAt time.Time,
) (RefreshToken, error) {
	refreshToken = r.hashToken(refreshToken)
	storedRefreshToken, err := r.queries.CreateRefreshToken(
		ctx,
		storage.CreateRefreshTokenParams{
			Hash:      refreshToken,
			SessionID: sessionId,
			ExpiresAt: expiresAt,
		},
	)
	if err != nil {
		return RefreshToken{}, err
	}
	return r.transformRefreshToken(storedRefreshToken), nil
}

func (r *DefaultTokenRepository) GetRefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error) {
	refreshToken = r.hashToken(refreshToken)
	storedRefreshToken, err := r.queries.GetRefreshTokenByHash(ctx, refreshToken)
	if err != nil {
		return RefreshToken{}, err
	}
	return r.transformRefreshToken(storedRefreshToken), nil
}

func (r *DefaultTokenRepository) GetAccessToken(ctx context.Context, accessToken string) (AccessToken, error) {
	accessToken = r.hashToken(accessToken)
	storedAccessToken, err := r.queries.GetAccessTokenByHash(ctx, accessToken)
	if err != nil {
		return AccessToken{}, err
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

func (r *DefaultTokenRepository) transformAccessToken(storedAccessToken storage.AccessToken) AccessToken {
	var data map[string]interface{}
	if err := json.Unmarshal(storedAccessToken.Data, &data); err != nil {
		data = make(map[string]interface{})
	}
	return AccessToken{
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

func (r *DefaultTokenRepository) transformRefreshToken(storedRefreshToken storage.RefreshToken) RefreshToken {
	return RefreshToken{
		Hash:      storedRefreshToken.Hash,
		SessionID: storedRefreshToken.SessionID,
		RevokedAt: storedRefreshToken.RevokedAt,
		ExpiresAt: storedRefreshToken.ExpiresAt,
	}
}

func (r *DefaultTokenRepository) hashToken(token string) string {
	if !r.config.HashTokens {
		return token
	}
	sha := sha1.New()
	sha.Write([]byte(token))
	return string(sha.Sum(nil))
}
