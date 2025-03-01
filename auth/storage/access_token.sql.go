// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: access_token.sql

package storage

import (
	"context"
	"time"

	uuid "github.com/gofrs/uuid"
)

const createAccessToken = `-- name: CreateAccessToken :one
INSERT INTO auth.access_token (hash, identity_id, session_id, user_id, roles, data, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING hash, identity_id, session_id, user_id, roles, data, revoked_at, expires_at, created_at`

type CreateAccessTokenParams struct {
	Hash       string    `db:"hash" json:"hash"`
	IdentityID uuid.UUID `db:"identity_id" json:"identityId"`
	SessionID  uuid.UUID `db:"session_id" json:"sessionId"`
	UserID     uuid.UUID `db:"user_id" json:"userId"`
	Roles      []string  `db:"roles" json:"roles"`
	Data       []byte    `db:"data" json:"data"`
	ExpiresAt  time.Time `db:"expires_at" json:"expiresAt"`
}

func (q *Queries) CreateAccessToken(ctx context.Context, arg CreateAccessTokenParams) (AccessToken, error) {
	row := q.db.QueryRow(ctx, createAccessToken,
		arg.Hash,
		arg.IdentityID,
		arg.SessionID,
		arg.UserID,
		arg.Roles,
		arg.Data,
		arg.ExpiresAt,
	)
	var i AccessToken
	err := row.Scan(
		&i.Hash,
		&i.IdentityID,
		&i.SessionID,
		&i.UserID,
		&i.Roles,
		&i.Data,
		&i.RevokedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getAccessTokenByHash = `-- name: GetAccessTokenByHash :one
SELECT hash, identity_id, session_id, user_id, roles, data, revoked_at, expires_at, created_at
FROM auth.access_token
WHERE hash = $1`

func (q *Queries) GetAccessTokenByHash(ctx context.Context, hash string) (AccessToken, error) {
	row := q.db.QueryRow(ctx, getAccessTokenByHash, hash)
	var i AccessToken
	err := row.Scan(
		&i.Hash,
		&i.IdentityID,
		&i.SessionID,
		&i.UserID,
		&i.Roles,
		&i.Data,
		&i.RevokedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getUserNotRevokedSessionIds = `-- name: GetUserNotRevokedSessionIds :many
SELECT session_id
FROM auth.access_token
WHERE revoked_at IS NULL
  AND user_id = $1`

func (q *Queries) GetUserNotRevokedSessionIds(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := q.db.Query(ctx, getUserNotRevokedSessionIds, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var session_id uuid.UUID
		if err := rows.Scan(&session_id); err != nil {
			return nil, err
		}
		items = append(items, session_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const revokeAccessToken = `-- name: RevokeAccessToken :exec
UPDATE auth.access_token
SET revoked_at = now()
WHERE hash = $1 AND revoked_at IS NULL`

func (q *Queries) RevokeAccessToken(ctx context.Context, hash string) error {
	_, err := q.db.Exec(ctx, revokeAccessToken, hash)
	return err
}

const revokeSessionAccessTokens = `-- name: RevokeSessionAccessTokens :exec
UPDATE auth.access_token
SET revoked_at = now()
WHERE session_id = $1 AND revoked_at IS NULL`

func (q *Queries) RevokeSessionAccessTokens(ctx context.Context, sessionID uuid.UUID) error {
	_, err := q.db.Exec(ctx, revokeSessionAccessTokens, sessionID)
	return err
}

const revokeUserAccessTokens = `-- name: RevokeUserAccessTokens :exec
UPDATE auth.access_token
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL`

func (q *Queries) RevokeUserAccessTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := q.db.Exec(ctx, revokeUserAccessTokens, userID)
	return err
}
