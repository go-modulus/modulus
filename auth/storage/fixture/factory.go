package fixture

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/go-modulus/modulus/auth/repository"
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
	id := uuid.Must(uuid.NewV6())
	hash := base64.StdEncoding.EncodeToString(id.Bytes())[:16]
	return NewCredentialFixture(
		f.db, storage.Credential{
			ID:             id,
			IdentityID:     uuid.Must(uuid.NewV6()),
			CredentialHash: hash,
			Type:           string(repository.CredentialTypePassword),
			ExpiredAt:      null.Time{},
			CreatedAt:      time.Now(),
		},
	)
}

func (f *FixturesFactory) Identity() *IdentityFixture {
	id := uuid.Must(uuid.NewV6())
	return NewIdentityFixture(
		f.db, storage.Identity{
			ID:        id,
			Identity:  "test" + id.String(),
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
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	hash := base64.StdEncoding.EncodeToString(bytes)[:32]
	return NewRefreshTokenFixture(
		f.db, storage.RefreshToken{
			Hash:      hash,
			SessionID: uuid.Must(uuid.NewV6()),
			RevokedAt: null.Time{},
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		},
	)
}

func (f *FixturesFactory) AccessToken() *AccessTokenFixture {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	hash := base64.StdEncoding.EncodeToString(bytes)[:32]
	return NewAccessTokenFixture(
		f.db, storage.AccessToken{
			Hash:       hash,
			IdentityID: uuid.Must(uuid.NewV6()),
			SessionID:  uuid.Must(uuid.NewV6()),
			UserID:     uuid.Must(uuid.NewV6()),
			Roles:      []string{},
			Data:       nil,
			RevokedAt:  null.Time{},
			ExpiresAt:  time.Now().Add(time.Hour),
			CreatedAt:  time.Now(),
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
