// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: refresh_token.sql

package storage

import (
	"context"
	"time"

	uuid "github.com/gofrs/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO auth.refresh_token (hash, session_id, data, expires_at)
VALUES ($1::text, $2::uuid, $3::jsonb, $4)
RETURNING hash, session_id, data, revoked_at, used_at, expires_at, created_at`

type CreateRefreshTokenParams struct {
	Hash      string    `db:"hash" json:"hash"`
	SessionID uuid.UUID `db:"session_id" json:"sessionId"`
	Data      []byte    `db:"data" json:"data"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRow(ctx, createRefreshToken,
		arg.Hash,
		arg.SessionID,
		arg.Data,
		arg.ExpiresAt,
	)
	var i RefreshToken
	err := row.Scan(
		&i.Hash,
		&i.SessionID,
		&i.Data,
		&i.RevokedAt,
		&i.UsedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getRefreshTokenByHash = `-- name: GetRefreshTokenByHash :one
SELECT hash, session_id, data, revoked_at, used_at, expires_at, created_at
FROM auth.refresh_token
WHERE hash = $1`

func (q *Queries) GetRefreshTokenByHash(ctx context.Context, hash string) (RefreshToken, error) {
	row := q.db.QueryRow(ctx, getRefreshTokenByHash, hash)
	var i RefreshToken
	err := row.Scan(
		&i.Hash,
		&i.SessionID,
		&i.Data,
		&i.RevokedAt,
		&i.UsedAt,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const revokeRefreshTokens = `-- name: RevokeRefreshTokens :exec
UPDATE auth.refresh_token
SET revoked_at = $2::timestamptz
WHERE session_id = $1 AND used_at IS NULL`

type RevokeRefreshTokensParams struct {
	SessionID uuid.UUID `db:"session_id" json:"sessionId"`
	RevokedAt time.Time `db:"revoked_at" json:"revokedAt"`
}

func (q *Queries) RevokeRefreshTokens(ctx context.Context, arg RevokeRefreshTokensParams) error {
	_, err := q.db.Exec(ctx, revokeRefreshTokens, arg.SessionID, arg.RevokedAt)
	return err
}

const useRefreshToken = `-- name: UseRefreshToken :exec
UPDATE auth.refresh_token
SET used_at = $2::timestamptz
WHERE hash = $1`

type UseRefreshTokenParams struct {
	Hash   string    `db:"hash" json:"hash"`
	UsedAt time.Time `db:"used_at" json:"usedAt"`
}

func (q *Queries) UseRefreshToken(ctx context.Context, arg UseRefreshTokenParams) error {
	_, err := q.db.Exec(ctx, useRefreshToken, arg.Hash, arg.UsedAt)
	return err
}
