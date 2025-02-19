package auth

import (
	"braces.dev/errtrace"
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"time"
)

var ErrTokenIsRevoked = errors.New("token is revoked")
var ErrTokenIsExpired = errors.New("token is expired")
var ErrCannotCreateAccessToken = errors.New("cannot create access token")
var ErrCannotCreateRefreshToken = errors.New("cannot create refresh token")

type PlainTokenAuthenticator struct {
	tokenRepository    TokenRepository
	identityRepository IdentityRepository
}

func NewPlainTokenAuthenticator(
	tokenRepository TokenRepository,
	identityRepository IdentityRepository,
) *PlainTokenAuthenticator {
	return &PlainTokenAuthenticator{
		tokenRepository:    tokenRepository,
		identityRepository: identityRepository,
	}
}

// Authenticate authenticates the user with the given token.
// It returns the performer of the authenticated user.
//
// Errors:
// * github.com/go-modulus/modulus/auth.ErrTokenIsRevoked - if the token is revoked.
// * github.com/go-modulus/modulus/auth.ErrTokenIsExpired - if the token is expired.
func (a *PlainTokenAuthenticator) Authenticate(ctx context.Context, token string) (Performer, error) {
	accessToken, err := a.tokenRepository.GetAccessToken(ctx, token)
	if err != nil {
		return Performer{}, err
	}

	if accessToken.RevokedAt.Valid {
		return Performer{}, ErrTokenIsRevoked
	}

	if accessToken.ExpiresAt.Before(time.Now()) {
		return Performer{}, ErrTokenIsExpired
	}

	return Performer{
		ID:        accessToken.UserID,
		SessionID: accessToken.SessionID,
	}, nil
}

// StartSession starts a new session for the given performer. It means creation the new pair of access and refresh tokens without revoking any existing tokens.
// It returns an access token and a refresh token.
// Errors:
// * github.com/go-modulus/modulus/auth.ErrCannotCreateAccessToken - if the access token cannot be created.
// * github.com/go-modulus/modulus/auth.ErrCannotCreateRefreshToken - if the refresh token cannot be created.
func (a *PlainTokenAuthenticator) StartSession(
	ctx context.Context,
	identityID uuid.UUID,
) (AccessToken, RefreshToken, error) {
	accessTokenStr, err := a.randomString(32)
	if err != nil {
		return AccessToken{}, RefreshToken{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateAccessToken, err))
	}

	refreshTokenStr, err := a.randomString(32)
	if err != nil {
		return AccessToken{}, RefreshToken{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateRefreshToken, err))
	}

	identity, err := a.identityRepository.GetById(ctx, identityID)
	if err != nil {
		return AccessToken{}, RefreshToken{}, errtrace.Wrap(err)
	}

	sessionID := uuid.Must(uuid.NewV6())
	accessToken, err := a.tokenRepository.CreateAccessToken(
		ctx,
		accessTokenStr,
		identityID,
		identity.UserID,
		identity.Roles,
		sessionID,
		map[string]interface{}{},
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return AccessToken{}, RefreshToken{}, err
	}

	refreshToken, err := a.tokenRepository.CreateRefreshToken(
		ctx,
		refreshTokenStr,
		sessionID,
		time.Now().Add(24*time.Hour),
	)
	if err != nil {
		return AccessToken{}, RefreshToken{}, err
	}

	return accessToken, refreshToken, nil
}

func (a *PlainTokenAuthenticator) randomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes)[:length], nil
}
