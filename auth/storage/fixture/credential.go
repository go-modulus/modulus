// Code generated by sqlc-fixture plugin for SQLc. DO NOT EDIT.

package fixture

import (
	"context"
	"github.com/go-modulus/modulus/auth/storage"
	uuid "github.com/gofrs/uuid"
	null "gopkg.in/guregu/null.v4"
	"testing"
	"time"
)

type CredentialFixture struct {
	entity storage.Credential
	db     storage.DBTX
}

func NewCredentialFixture(db storage.DBTX, defaultEntity storage.Credential) *CredentialFixture {
	return &CredentialFixture{
		db:     db,
		entity: defaultEntity,
	}
}

func (f *CredentialFixture) Hash(hash string) *CredentialFixture {
	c := f.clone()
	c.entity.Hash = hash
	return c
}

func (f *CredentialFixture) IdentityID(identityID uuid.UUID) *CredentialFixture {
	c := f.clone()
	c.entity.IdentityID = identityID
	return c
}

func (f *CredentialFixture) Type(typ string) *CredentialFixture {
	c := f.clone()
	c.entity.Type = typ
	return c
}

func (f *CredentialFixture) ExpiredAt(expiredAt null.Time) *CredentialFixture {
	c := f.clone()
	c.entity.ExpiredAt = expiredAt
	return c
}

func (f *CredentialFixture) CreatedAt(createdAt time.Time) *CredentialFixture {
	c := f.clone()
	c.entity.CreatedAt = createdAt
	return c
}

func (f *CredentialFixture) clone() *CredentialFixture {
	return &CredentialFixture{
		db:     f.db,
		entity: f.entity,
	}
}

func (f *CredentialFixture) save(ctx context.Context) error {
	query := `INSERT INTO auth.credential
            (hash, identity_id, type, expired_at, created_at)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING hash, identity_id, type, expired_at, created_at
        `
	row := f.db.QueryRow(ctx, query,
		f.entity.Hash,
		f.entity.IdentityID,
		f.entity.Type,
		f.entity.ExpiredAt,
		f.entity.CreatedAt,
	)
	err := row.Scan(
		&f.entity.Hash,
		&f.entity.IdentityID,
		&f.entity.Type,
		&f.entity.ExpiredAt,
		&f.entity.CreatedAt,
	)
	return err
}

func (f *CredentialFixture) GetEntity() storage.Credential {
	return f.entity
}

func (f *CredentialFixture) Create(tb testing.TB) *CredentialFixture {
	err := f.save(context.Background())
	if err != nil {
		tb.Fatalf("failed to create Credential: %v", err)
	}
	f.Cleanup(tb)
	c := f.clone()
	return c
}

// Cleanup calls testing.TB.Cleanup() function with providing a callback inside it.
// This callback will delete a record from the table by primary key when test will be finished.
func (f *CredentialFixture) Cleanup(tb testing.TB) *CredentialFixture {
	tb.Cleanup(
		func() {
			query := `DELETE FROM auth.credential WHERE hash = $1`
			_, err := f.db.Exec(context.Background(), query, f.entity.Hash)

			if err != nil {
				tb.Fatalf("failed to cleanup Credential: %v", err)
			}
		})

	return f
}

func (f *CredentialFixture) PullUpdates(tb testing.TB) *CredentialFixture {
	c := f.clone()
	ctx := context.Background()
	query := `SELECT * FROM auth.credential WHERE hash = $1`
	row := f.db.QueryRow(ctx, query,
		c.entity.Hash,
	)

	err := row.Scan(
		&c.entity.Hash,
		&c.entity.IdentityID,
		&c.entity.Type,
		&c.entity.ExpiredAt,
		&c.entity.CreatedAt,
	)
	if err != nil {
		tb.Fatalf("failed to actualize data Credential: %v", err)
	}
	return c
}

func (f *CredentialFixture) PushUpdates(tb testing.TB) *CredentialFixture {
	c := f.clone()
	query := `
        UPDATE auth.credential SET 
            identity_id = $2,
            type = $3,
            expired_at = $4,
            created_at = $5
        WHERE hash = $1
        `
	_, err := f.db.Exec(
		context.Background(),
		query,
		f.entity.Hash,
		f.entity.IdentityID,
		f.entity.Type,
		f.entity.ExpiredAt,
		f.entity.CreatedAt,
	)
	if err != nil {
		tb.Fatalf("failed to push the data Credential: %v", err)
	}
	return c
}
