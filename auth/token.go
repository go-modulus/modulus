package auth

import (
	"braces.dev/errtrace"
	"context"
	"encoding/json"
	"github.com/go-modulus/modulus/auth/hash"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrTokenNotExist = errtrace.New("token does not exist")

type AccessToken struct {
	Token      null.String            `json:"token"`
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
	Token     null.String `json:"token"`
	Hash      string      `json:"hash"`
	SessionID uuid.UUID   `json:"sessionId"`
	RevokedAt null.Time   `json:"revokedAt"`
	ExpiresAt time.Time   `json:"expiresAt"`
}

type TokenRepository interface {
	// CreateAccessToken creates an access token.
	// It returns the created access token.
	//
	// Errors:
	// * github.com/go-modulus/modulus/auth.ErrCannotCreateAccessToken - if the access token cannot be created.
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
	// CreateRefreshToken creates a refresh token.
	// It returns the created refresh token.
	//
	// Errors:
	// * github.com/go-modulus/modulus/auth.ErrCannotCreateRefreshToken - if the refresh token cannot be created.
	CreateRefreshToken(
		ctx context.Context,
		refreshToken string,
		sessionId uuid.UUID,
		expiresAt time.Time,
	) (RefreshToken, error)
	// GetRefreshToken returns the refresh token by the given token.
	//
	// Errors:
	// * github.com/go-modulus/modulus/auth.ErrTokenNotExist - if the token does not exist.
	GetRefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error)
	// GetAccessToken returns the access token by the given token.
	//
	// Errors:
	// * github.com/go-modulus/modulus/auth.ErrTokenNotExist - if the token does not exist.
	GetAccessToken(ctx context.Context, accessToken string) (AccessToken, error)
	// RevokeAccessToken revokes the access token by the given token. Be careful, with token not its hash, stored in DB
	RevokeAccessToken(ctx context.Context, accessToken string) error
	// RevokeRefreshToken revokes the refresh token by the given token. Be careful, with token not its hash, stored in DB
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	// RevokeSessionTokens revokes all tokens of the session by the given session ID.
	RevokeSessionTokens(ctx context.Context, sessionId uuid.UUID) error
	// RevokeUserTokens revokes all tokens of the user by the given user ID.
	RevokeUserTokens(ctx context.Context, userId uuid.UUID) error
}

type DefaultTokenRepository struct {
	queries      *storage.Queries
	config       ModuleConfig
	hashStrategy hash.TokenHashStrategy
}

func NewDefaultTokenRepository(
	db *pgxpool.Pool,
	config ModuleConfig,
	hashStrategy hash.TokenHashStrategy,
) TokenRepository {
	return &DefaultTokenRepository{
		queries:      storage.New(db),
		config:       config,
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
) (AccessToken, error) {
	accessToken = r.hashToken(accessToken)

	if data == nil {
		data = make(map[string]interface{})
	}
	dataJson, err := json.Marshal(data)
	if err != nil {
		return AccessToken{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateAccessToken, err))
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
		return RefreshToken{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateRefreshToken, err))
	}
	return r.transformRefreshToken(storedRefreshToken), nil
}

func (r *DefaultTokenRepository) GetRefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error) {
	refreshToken = r.hashToken(refreshToken)
	storedRefreshToken, err := r.queries.GetRefreshTokenByHash(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshToken{}, errtrace.Wrap(ErrTokenNotExist)
		}
		return RefreshToken{}, errtrace.Wrap(err)
	}
	return r.transformRefreshToken(storedRefreshToken), nil
}

func (r *DefaultTokenRepository) GetAccessToken(ctx context.Context, accessToken string) (AccessToken, error) {
	accessToken = r.hashToken(accessToken)
	storedAccessToken, err := r.queries.GetAccessTokenByHash(ctx, accessToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AccessToken{}, errtrace.Wrap(ErrTokenNotExist)
		}
		return AccessToken{}, errtrace.Wrap(err)
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
	return r.hashStrategy.Hash(token)
}
