package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrTokenNotExist = errors.New("token does not exist")
var ErrCannotCreateAccessToken = errors.New("cannot create access token")
var ErrCannotCreateRefreshToken = errors.New("cannot create refresh token")

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
	Token      null.String `json:"token"`
	Hash       string      `json:"hash"`
	IdentityID uuid.UUID   `json:"identityId"`
	SessionID  uuid.UUID   `json:"sessionId"`
	RevokedAt  null.Time   `json:"revokedAt"`
	ExpiresAt  time.Time   `json:"expiresAt"`
}

type TokenRepository interface {
	// CreateAccessToken creates an access token.
	// It returns the created access token.
	//
	// Errors:
	// * ErrCannotCreateAccessToken - if the access token cannot be created.
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
	// * ErrCannotCreateRefreshToken - if the refresh token cannot be created.
	CreateRefreshToken(
		ctx context.Context,
		refreshToken string,
		sessionID uuid.UUID,
		identityID uuid.UUID,
		expiresAt time.Time,
	) (RefreshToken, error)
	// GetRefreshToken returns the refresh token by the given token.
	//
	// Errors:
	// * ErrTokenNotExist - if the token does not exist.
	GetRefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error)
	// GetAccessToken returns the access token by the given token.
	//
	// Errors:
	// * ErrTokenNotExist - if the token does not exist.
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
