package auth

import (
	"braces.dev/errtrace"
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

var ErrTokenIsRevoked = errors.New("token is revoked")
var ErrTokenIsExpired = errors.New("token is expired")
var ErrCannotCreateAccessToken = errors.New("cannot create access token")
var ErrCannotCreateRefreshToken = errors.New("cannot create refresh token")

type TokenPair struct {
	AccessToken  AccessToken
	RefreshToken RefreshToken
}

type PlainTokenAuthenticator struct {
	tokenRepository    TokenRepository
	identityRepository IdentityRepository
	config             ModuleConfig
}

func NewPlainTokenAuthenticator(
	tokenRepository TokenRepository,
	identityRepository IdentityRepository,
	config ModuleConfig,
) *PlainTokenAuthenticator {
	return &PlainTokenAuthenticator{
		tokenRepository:    tokenRepository,
		identityRepository: identityRepository,
		config:             config,
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
		Roles:     accessToken.Roles,
	}, nil
}

// IssueTokens starts a new session for the given performer. It means creation the new pair of access and refresh tokens without revoking any existing tokens.
// It returns an access token and a refresh token.
//
// The additionalData parameter is used to store additional data in the access token. For example, you can store the IP address of the user.
//
// Errors:
// * github.com/go-modulus/modulus/auth.ErrCannotCreateAccessToken - if the access token cannot be created.
// * github.com/go-modulus/modulus/auth.ErrCannotCreateRefreshToken - if the refresh token cannot be created.
func (a *PlainTokenAuthenticator) IssueTokens(
	ctx context.Context,
	identityID uuid.UUID,
	additionalData map[string]interface{},
) (TokenPair, error) {
	accessTokenStr, err := a.randomString(32)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateAccessToken, err))
	}

	refreshTokenStr, err := a.randomString(32)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(errors.WrapCause(ErrCannotCreateRefreshToken, err))
	}

	identity, err := a.identityRepository.GetById(ctx, identityID)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

	sessionID := uuid.Must(uuid.NewV6())
	accessToken, err := a.tokenRepository.CreateAccessToken(
		ctx,
		accessTokenStr,
		identityID,
		identity.UserID,
		identity.Roles,
		sessionID,
		additionalData,
		time.Now().Add(a.config.AccessTokenTTL),
	)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

	refreshToken, err := a.tokenRepository.CreateRefreshToken(
		ctx,
		refreshTokenStr,
		sessionID,
		time.Now().Add(a.config.RefreshTokenTTL),
	)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken.Token = null.StringFrom(accessTokenStr)
	refreshToken.Token = null.StringFrom(refreshTokenStr)

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *PlainTokenAuthenticator) randomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes)[:length], nil
}
