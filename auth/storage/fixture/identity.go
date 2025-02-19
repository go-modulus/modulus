// Code generated by sqlc-fixture plugin for SQLc. DO NOT EDIT.

package fixture

import (
	"context"
	"github.com/go-modulus/modulus/auth/storage"
	uuid "github.com/gofrs/uuid"
	"testing"
	"time"
)

type IdentityFixture struct {
	entity storage.Identity
	db     storage.DBTX
}

func NewIdentityFixture(db storage.DBTX, defaultEntity storage.Identity) *IdentityFixture {
	return &IdentityFixture{
		db:     db,
		entity: defaultEntity,
	}
}

func (f *IdentityFixture) ID(iD uuid.UUID) *IdentityFixture {
	c := f.clone()
	c.entity.ID = iD
	return c
}

func (f *IdentityFixture) Identity(identity string) *IdentityFixture {
	c := f.clone()
	c.entity.Identity = identity
	return c
}

func (f *IdentityFixture) UserID(userID uuid.UUID) *IdentityFixture {
	c := f.clone()
	c.entity.UserID = userID
	return c
}

func (f *IdentityFixture) Status(status storage.IdentityStatus) *IdentityFixture {
	c := f.clone()
	c.entity.Status = status
	return c
}

func (f *IdentityFixture) Data(data []byte) *IdentityFixture {
	c := f.clone()
	c.entity.Data = data
	return c
}

func (f *IdentityFixture) Roles(roles string) *IdentityFixture {
	c := f.clone()
	c.entity.Roles = roles
	return c
}

func (f *IdentityFixture) UpdatedAt(updatedAt time.Time) *IdentityFixture {
	c := f.clone()
	c.entity.UpdatedAt = updatedAt
	return c
}

func (f *IdentityFixture) CreatedAt(createdAt time.Time) *IdentityFixture {
	c := f.clone()
	c.entity.CreatedAt = createdAt
	return c
}

func (f *IdentityFixture) clone() *IdentityFixture {
	return &IdentityFixture{
		db:     f.db,
		entity: f.entity,
	}
}

func (f *IdentityFixture) save(ctx context.Context) error {
	query := `INSERT INTO auth.identity
            (id, identity, user_id, status, data, roles, updated_at, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            RETURNING id, identity, user_id, status, data, roles, updated_at, created_at
        `
	row := f.db.QueryRow(ctx, query,
		f.entity.ID,
		f.entity.Identity,
		f.entity.UserID,
		f.entity.Status,
		f.entity.Data,
		f.entity.Roles,
		f.entity.UpdatedAt,
		f.entity.CreatedAt,
	)
	err := row.Scan(
		&f.entity.ID,
		&f.entity.Identity,
		&f.entity.UserID,
		&f.entity.Status,
		&f.entity.Data,
		&f.entity.Roles,
		&f.entity.UpdatedAt,
		&f.entity.CreatedAt,
	)
	return err
}

func (f *IdentityFixture) GetEntity() storage.Identity {
	return f.entity
}

func (f *IdentityFixture) Create(tb testing.TB) *IdentityFixture {
	err := f.save(context.Background())
	if err != nil {
		tb.Fatalf("failed to create Identity: %v", err)
	}
	f.Cleanup(tb)
	c := f.clone()
	return c
}

// Cleanup calls testing.TB.Cleanup() function with providing a callback inside it.
// This callback will delete a record from the table by primary key when test will be finished.
func (f *IdentityFixture) Cleanup(tb testing.TB) *IdentityFixture {
	tb.Cleanup(
		func() {
			query := `DELETE FROM auth.identity WHERE id = $1`
			_, err := f.db.Exec(context.Background(), query, f.entity.ID)

			if err != nil {
				tb.Fatalf("failed to cleanup Identity: %v", err)
			}
		})

	return f
}

func (f *IdentityFixture) PullUpdates(tb testing.TB) *IdentityFixture {
	c := f.clone()
	ctx := context.Background()
	query := `SELECT * FROM auth.identity WHERE id = $1`
	row := f.db.QueryRow(ctx, query,
		c.entity.ID,
	)

	err := row.Scan(
		&c.entity.ID,
		&c.entity.Identity,
		&c.entity.UserID,
		&c.entity.Status,
		&c.entity.Data,
		&c.entity.Roles,
		&c.entity.UpdatedAt,
		&c.entity.CreatedAt,
	)
	if err != nil {
		tb.Fatalf("failed to actualize data Identity: %v", err)
	}
	return c
}

func (f *IdentityFixture) PushUpdates(tb testing.TB) *IdentityFixture {
	c := f.clone()
	query := `
        UPDATE auth.identity SET 
            identity = $2,
            user_id = $3,
            status = $4,
            data = $5,
            roles = $6,
            updated_at = $7,
            created_at = $8
        WHERE id = $1
        `
	_, err := f.db.Exec(
		context.Background(),
		query,
		f.entity.ID,
		f.entity.Identity,
		f.entity.UserID,
		f.entity.Status,
		f.entity.Data,
		f.entity.Roles,
		f.entity.UpdatedAt,
		f.entity.CreatedAt,
	)
	if err != nil {
		tb.Fatalf("failed to push the data Identity: %v", err)
	}
	return c
}
