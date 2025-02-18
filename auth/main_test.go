package auth_test

import (
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/auth/storage/fixture"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"testing"
)

var (
	passwordAuth      *auth.PasswordAuthenticator
	credentialFixture *fixture.CredentialFixture
	identityFixture   *fixture.IdentityFixture
)

func TestMain(m *testing.M) {
	test.LoadEnv("..")
	mod := auth.NewModule().
		AddFxOptions(fixture.FxProvide()).
		AddProviders(
			func(db *pgxpool.Pool) storage.DBTX {
				return db
			},
		)

	test.TestMain(
		m,
		module.BuildFx(mod),
		fx.Populate(
			&passwordAuth,
			&credentialFixture,
			&identityFixture,
		),
	)
}
