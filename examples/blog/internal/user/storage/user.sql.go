// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: user.sql

package storage

import (
	"context"

	uuid "github.com/gofrs/uuid"
)

const findUserByEmail = `-- name: FindUserByEmail :one
SELECT id, email, name, created_at, updated_at FROM "user"."user"
WHERE email = $1::text`

func (q *Queries) FindUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, findUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const registerUser = `-- name: RegisterUser :one
INSERT INTO "user"."user" (id, email, name)
VALUES ($1::uuid, $2::text, $3::text)
RETURNING id, email, name, created_at, updated_at`

type RegisterUserParams struct {
	ID    uuid.UUID `db:"id" json:"id"`
	Email string    `db:"email" json:"email"`
	Name  string    `db:"name" json:"name"`
}

func (q *Queries) RegisterUser(ctx context.Context, arg RegisterUserParams) (User, error) {
	row := q.db.QueryRow(ctx, registerUser, arg.ID, arg.Email, arg.Name)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
