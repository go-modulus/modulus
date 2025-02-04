// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: identity.sql

package storage

import (
	"context"

	uuid "github.com/gofrs/uuid"
)

const createIdentity = `-- name: CreateIdentity :one
insert into "auth"."identity"
    (id, identity, user_id, "data")
values ($1::uuid, $2::text, $3::uuid, $4::jsonb)
RETURNING id, identity, user_id, status, data, updated_at, created_at`

type CreateIdentityParams struct {
	ID       uuid.UUID `db:"id" json:"id"`
	Identity string    `db:"identity" json:"identity"`
	UserID   uuid.UUID `db:"user_id" json:"userId"`
	Data     []byte    `db:"data" json:"data"`
}

func (q *Queries) CreateIdentity(ctx context.Context, arg CreateIdentityParams) (Identity, error) {
	row := q.db.QueryRow(ctx, createIdentity,
		arg.ID,
		arg.Identity,
		arg.UserID,
		arg.Data,
	)
	var i Identity
	err := row.Scan(
		&i.ID,
		&i.Identity,
		&i.UserID,
		&i.Status,
		&i.Data,
		&i.UpdatedAt,
		&i.CreatedAt,
	)
	return i, err
}

const deleteIdentity = `-- name: DeleteIdentity :exec
delete from "auth"."identity"
where id = $1`

func (q *Queries) DeleteIdentity(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteIdentity, id)
	return err
}

const findIdentity = `-- name: FindIdentity :one
select id, identity, user_id, status, data, updated_at, created_at
from "auth"."identity"
where identity = $1::text`

func (q *Queries) FindIdentity(ctx context.Context, identity string) (Identity, error) {
	row := q.db.QueryRow(ctx, findIdentity, identity)
	var i Identity
	err := row.Scan(
		&i.ID,
		&i.Identity,
		&i.UserID,
		&i.Status,
		&i.Data,
		&i.UpdatedAt,
		&i.CreatedAt,
	)
	return i, err
}

const findUserIdentities = `-- name: FindUserIdentities :many
select id, identity, user_id, status, data, updated_at, created_at
from "auth"."identity"
where user_id = $1::uuid`

func (q *Queries) FindUserIdentities(ctx context.Context, userID uuid.UUID) ([]Identity, error) {
	rows, err := q.db.Query(ctx, findUserIdentities, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Identity
	for rows.Next() {
		var i Identity
		if err := rows.Scan(
			&i.ID,
			&i.Identity,
			&i.UserID,
			&i.Status,
			&i.Data,
			&i.UpdatedAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
