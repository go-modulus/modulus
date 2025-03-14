package repository

import (
	"context"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type ExpirationTokenType string

const (
	AccessTokenType  ExpirationTokenType = "access"
	RefreshTokenType ExpirationTokenType = "refresh"
	BothTokenType    ExpirationTokenType = "both"
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

	// ExpireTokens makes the valid tokens of the given session expired.
	// It returns an error if the operation failed.
	// Params:
	// * sessionId - the session where we want to expire the tokens.
	// * lag - the parameter is used to make the tokens expired in the future. It is helpful to avoid race conditions in the token refreshing.
	// 		   We make tokens expired in several seconds to get the time for frontend to refresh the tokens in local storage and get the responses from simultaneous requests to the backend without error.
	// * tokenType - the type of tokens that we want to expire.
	ExpireTokens(
		ctx context.Context,
		sessionId uuid.UUID,
		lag time.Duration,
		tokenType ExpirationTokenType,
	) error

	// RevokeSessionTokens revokes all tokens of the session by the given session ID.
	// It can be used from the user settings to log out from some devices.
	RevokeSessionTokens(ctx context.Context, sessionId uuid.UUID) error

	// RevokeUserTokens revokes all tokens of the user by the given user ID.
	// It can be used from the user settings to log out from all devices.
	RevokeUserTokens(ctx context.Context, userId uuid.UUID) error
}
