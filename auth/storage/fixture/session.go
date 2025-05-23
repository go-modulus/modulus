// Code generated by sqlc-fixture plugin for SQLc. DO NOT EDIT.

package fixture

import (
	"context"
	"github.com/go-modulus/modulus/auth/storage"
	uuid "github.com/gofrs/uuid"
	"testing"
	"time"
)

type SessionFixture struct {
	entity storage.Session
	db     storage.DBTX
}

func NewSessionFixture(db storage.DBTX, defaultEntity storage.Session) *SessionFixture {
	return &SessionFixture{
		db:     db,
		entity: defaultEntity,
	}
}

func (f *SessionFixture) ID(iD uuid.UUID) *SessionFixture {
	c := f.clone()
	c.entity.ID = iD
	return c
}

func (f *SessionFixture) AccountID(accountID uuid.UUID) *SessionFixture {
	c := f.clone()
	c.entity.AccountID = accountID
	return c
}

func (f *SessionFixture) IdentityID(identityID uuid.UUID) *SessionFixture {
	c := f.clone()
	c.entity.IdentityID = identityID
	return c
}

func (f *SessionFixture) Data(data []byte) *SessionFixture {
	c := f.clone()
	c.entity.Data = data
	return c
}

func (f *SessionFixture) ExpiresAt(expiresAt time.Time) *SessionFixture {
	c := f.clone()
	c.entity.ExpiresAt = expiresAt
	return c
}

func (f *SessionFixture) CreatedAt(createdAt time.Time) *SessionFixture {
	c := f.clone()
	c.entity.CreatedAt = createdAt
	return c
}

func (f *SessionFixture) clone() *SessionFixture {
	return &SessionFixture{
		db:     f.db,
		entity: f.entity,
	}
}

func (f *SessionFixture) save(ctx context.Context) error {
	query := `INSERT INTO "auth"."session"
            ("id", "account_id", "identity_id", "data", "expires_at", "created_at")
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING "id", "account_id", "identity_id", "data", "expires_at", "created_at"
        `
	row := f.db.QueryRow(ctx, query,
		f.entity.ID,
		f.entity.AccountID,
		f.entity.IdentityID,
		f.entity.Data,
		f.entity.ExpiresAt,
		f.entity.CreatedAt,
	)
	err := row.Scan(
		&f.entity.ID,
		&f.entity.AccountID,
		&f.entity.IdentityID,
		&f.entity.Data,
		&f.entity.ExpiresAt,
		&f.entity.CreatedAt,
	)
	return err
}

func (f *SessionFixture) GetEntity() storage.Session {
	return f.entity
}

func (f *SessionFixture) Create(tb testing.TB) *SessionFixture {
	err := f.save(context.Background())
	if err != nil {
		tb.Fatalf("failed to create Session: %v", err)
	}
	f.Cleanup(tb)
	c := f.clone()
	return c
}

// Cleanup calls testing.TB.Cleanup() function with providing a callback inside it.
// This callback will delete a record from the table by primary key when test will be finished.
func (f *SessionFixture) Cleanup(tb testing.TB) *SessionFixture {
	tb.Cleanup(
		func() {
			query := `DELETE FROM "auth"."session" WHERE id = $1`
			_, err := f.db.Exec(context.Background(), query, f.entity.ID)

			if err != nil {
				tb.Fatalf("failed to cleanup Session: %v", err)
			}
		})

	return f
}

func (f *SessionFixture) PullUpdates(tb testing.TB) *SessionFixture {
	c := f.clone()
	ctx := context.Background()
	query := `SELECT "id", "account_id", "identity_id", "data", "expires_at", "created_at" FROM "auth"."session" WHERE id = $1`
	row := f.db.QueryRow(ctx, query,
		c.entity.ID,
	)

	err := row.Scan(
		&c.entity.ID,
		&c.entity.AccountID,
		&c.entity.IdentityID,
		&c.entity.Data,
		&c.entity.ExpiresAt,
		&c.entity.CreatedAt,
	)
	if err != nil {
		tb.Fatalf("failed to actualize data Session: %v", err)
	}
	return c
}

func (f *SessionFixture) PushUpdates(tb testing.TB) *SessionFixture {
	c := f.clone()
	query := `
        UPDATE "auth"."session" SET 
            "account_id" = $2,
            "identity_id" = $3,
            "data" = $4,
            "expires_at" = $5,
            "created_at" = $6
        WHERE "id" = $1
        `
	_, err := f.db.Exec(
		context.Background(),
		query,
		f.entity.ID,
		f.entity.AccountID,
		f.entity.IdentityID,
		f.entity.Data,
		f.entity.ExpiresAt,
		f.entity.CreatedAt,
	)
	if err != nil {
		tb.Fatalf("failed to push the data Session: %v", err)
	}
	return c
}
