package fixture

import (
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/gofrs/uuid"
	"go.uber.org/fx"
	"gopkg.in/guregu/null.v4"
	"time"
)

func FxProvide() fx.Option {
	return fx.Provide(
		func(db storage.DBTX) *CredentialFixture {
			return NewCredentialFixture(
				db, storage.Credential{
					ID:             uuid.Must(uuid.NewV6()),
					IdentityID:     uuid.Must(uuid.NewV6()),
					CredentialHash: "test",
					Type:           string(auth.CredentialTypePassword),
					ExpiredAt:      null.Time{},
					CreatedAt:      time.Now(),
				},
			)
		},
		func(db storage.DBTX) *IdentityFixture {
			return NewIdentityFixture(
				db, storage.Identity{
					ID:        uuid.Must(uuid.NewV6()),
					Identity:  "test",
					UserID:    uuid.Must(uuid.NewV6()),
					Status:    storage.IdentityStatusActive,
					Data:      nil,
					UpdatedAt: time.Now(),
					CreatedAt: time.Now(),
				},
			)
		},
		func(db storage.DBTX) *RefreshTokenFixture {
			return NewRefreshTokenFixture(
				db, storage.RefreshToken{
					Hash:      "test",
					SessionID: uuid.Must(uuid.NewV6()),
					Data:      nil,
					RevokedAt: null.Time{},
					UsedAt:    null.Time{},
					ExpiresAt: time.Now().Add(time.Hour),
					CreatedAt: time.Now(),
				},
			)
		},
		func(db storage.DBTX) *SessionFixture {
			return NewSessionFixture(
				db, storage.Session{
					ID:         uuid.Must(uuid.NewV6()),
					UserID:     uuid.Must(uuid.NewV6()),
					IdentityID: uuid.Must(uuid.NewV6()),
					Data:       nil,
					ExpiresAt:  time.Now().Add(time.Hour),
					CreatedAt:  time.Now(),
				},
			)
		},
	)
}
