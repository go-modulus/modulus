package action_test

import (
	"github.com/go-modulus/modulus/auth/providers/email"
	"github.com/go-modulus/modulus/auth/providers/email/action"
	"github.com/go-modulus/modulus/auth/providers/email/action/mocks"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/auth/storage"
	"github.com/go-modulus/modulus/auth/storage/fixture"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"testing"
)

var (
	register             *action.Register
	authFixture          *fixture.FixturesFactory
	creatorMock          *mocks.MockUserCreator
	identityRepository   repository.IdentityRepository
	credentialRepository repository.CredentialRepository
	accountRepository    repository.AccountRepository
	resetPassword        *action.ResetPassword
	changePassword       *action.ChangePassword
	login                *action.Login
)

func createModule() *module.Module {
	return email.NewModule().
		AddProviders(
			func(db *pgxpool.Pool) storage.DBTX {
				return db
			},
			fixture.NewFixturesFactory,
		)
}

func TestMain(m *testing.M) {
	test.LoadEnv("../../../..")

	mod := createModule()
	mod.AddProviders(
		func() *mocks.MockUserCreator {
			creatorMock = &mocks.MockUserCreator{}

			return creatorMock
		},
	)

	mod = email.OverrideUserCreator[*mocks.MockUserCreator](
		mod,
	)

	test.TestMain(
		m,
		module.BuildFx(mod),
		fx.Populate(
			&register,
			&authFixture,
			&identityRepository,
			&credentialRepository,
			&accountRepository,
			&resetPassword,
			&changePassword,
			&login,
		),
	)
}
