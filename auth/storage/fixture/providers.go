package fixture

import (
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type FixturesFactory struct {
	db storage.DBTX
}

func NewFixturesFactory(db storage.DBTX) *FixturesFactory {
	return &FixturesFactory{
		db: db,
	}
}

func (f *FixturesFactory) Credential() *CredentialFixture {
	return NewCredentialFixture(
		f.db, storage.Credential{
			ID:             uuid.Must(uuid.NewV6()),
			IdentityID:     uuid.Must(uuid.NewV6()),
			CredentialHash: "test",
			Type:           string(auth.CredentialTypePassword),
			ExpiredAt:      null.Time{},
			CreatedAt:      time.Now(),
		},
	)
}

func (f *FixturesFactory) Identity() *IdentityFixture {
	return NewIdentityFixture(
		f.db, storage.Identity{
			ID:        uuid.Must(uuid.NewV6()),
			Identity:  "test",
			UserID:    uuid.Must(uuid.NewV6()),
			Roles:     []string{},
			Status:    storage.IdentityStatusActive,
			Data:      nil,
			UpdatedAt: time.Now(),
			CreatedAt: time.Now(),
		},
	)
}

func (f *FixturesFactory) RefreshToken() *RefreshTokenFixture {
	return NewRefreshTokenFixture(
		f.db, storage.RefreshToken{
			Hash:      "test",
			SessionID: uuid.Must(uuid.NewV6()),
			RevokedAt: null.Time{},
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		},
	)
}

func (f *FixturesFactory) Session() *SessionFixture {
	return NewSessionFixture(
		f.db, storage.Session{
			ID:         uuid.Must(uuid.NewV6()),
			UserID:     uuid.Must(uuid.NewV6()),
			IdentityID: uuid.Must(uuid.NewV6()),
			Data:       nil,
			ExpiresAt:  time.Now().Add(time.Hour),
			CreatedAt:  time.Now(),
		},
	)
}
