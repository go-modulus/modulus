package action_test

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	authEmail "github.com/go-modulus/modulus/auth/providers/email"
	"github.com/go-modulus/modulus/auth/providers/email/action"
	"github.com/go-modulus/modulus/auth/providers/email/action/mocks"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gopkg.in/guregu/null.v4"
	"testing"
	"time"
)

func TestResetPassword_Request(t *testing.T) {
	t.Parallel()
	t.Run(
		"send reset password for a user registered via email", func(t *testing.T) {
			t.Parallel()
			email := gofakeit.Email()
			authFixture.Identity().
				Type(string(repository.IdentityTypeEmail)).Identity(email).Create(t)

			mod := createModule()

			senderMock := mocks.NewMockMailSender(t)

			senderMock.EXPECT().ResetPasswordEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(
				nil,
			)
			mod.AddProviders(
				func() *mocks.MockMailSender {
					return senderMock
				},
			)
			mod = authEmail.OverrideMailSender[*mocks.MockMailSender](mod)

			var resetPassword *action.ResetPassword
			err := test.Invoke(
				module.BuildFx(mod),
				fx.Populate(&resetPassword),
			)
			require.NoError(t, err)

			// Create a reset password request
			rp, err := resetPassword.Request(
				context.Background(),
				email,
			)

			savedRp := authFixture.ResetPasswordRequest().ID(rp.ID).PullUpdates(t).Cleanup(t).GetEntity()

			t.Log("Given an identity registered via email")
			t.Logf("When I request a reset password for %s", email)
			t.Log("	Then the reset password entity should be saved")
			require.NoError(t, err)
			require.Equal(t, savedRp.ID, rp.ID)
			require.NotEmpty(t, savedRp.Token)
			t.Log("	And email should be sent")
			senderMock.AssertExpectations(t)
		},
	)

	t.Run(
		"send reset password email after cooldown period for a user registered via email", func(t *testing.T) {
			t.Parallel()
			email := gofakeit.Email()
			// prepare data
			account := authFixture.Account().Create(t).GetEntity()
			authFixture.Identity().
				AccountID(account.ID).
				Type(string(repository.IdentityTypeEmail)).
				Identity(email).
				Create(t)
			sentRequest := authFixture.ResetPasswordRequest().
				AccountID(account.ID).
				LastSendAt(null.TimeFrom(time.Now().Add(-6 * time.Minute))).
				Create(t).
				GetEntity()

			// prepare mocks
			mod := createModule()

			senderMock := mocks.NewMockMailSender(t)

			senderMock.EXPECT().ResetPasswordEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(
				nil,
			)
			mod.AddProviders(
				func() *mocks.MockMailSender {
					return senderMock
				},
			)
			mod = authEmail.OverrideMailSender[*mocks.MockMailSender](mod)

			var resetPassword *action.ResetPassword
			err := test.Invoke(
				module.BuildFx(mod),
				fx.Populate(&resetPassword),
			)
			require.NoError(t, err)

			// Create a reset password request
			rp, err := resetPassword.Request(
				context.Background(),
				email,
			)

			savedRp := authFixture.ResetPasswordRequest().ID(rp.ID).PullUpdates(t).Cleanup(t).GetEntity()

			t.Log("Given an identity registered via email")
			t.Log("Given a reset password request sent before a cooldown period")
			t.Logf("When I request a reset password for %s", email)
			t.Log("	Then the reset password entity should be saved")
			require.NoError(t, err)
			require.Equal(t, savedRp.ID, rp.ID)
			require.NotEmpty(t, savedRp.Token)
			t.Log("	And email should be sent")
			senderMock.AssertExpectations(t)
			t.Log("	And cooldown period should be reset")
			require.NotEqual(t, savedRp.LastSendAt.Time, time.Time{})
			require.Equal(t, sentRequest.ID.String(), rp.ID.String())
		},
	)

	t.Run(
		"send reset password for a user registered via another provider", func(t *testing.T) {
			t.Parallel()
			email := gofakeit.Email()
			// prepare data
			account := authFixture.Account().Create(t).GetEntity()
			authFixture.Identity().
				AccountID(account.ID).
				Type("google").
				Identity("1234456666").
				Create(t)

			// create mocks
			mod := createModule()
			senderMock := createSenderMockSuccess(t, email, mod)
			verifierMock := createVerifiedEmailCheckerMockSuccess(t, email, account.ID, mod)
			mod = authEmail.OverrideMailSender[*mocks.MockMailSender](mod)
			mod = authEmail.OverrideVerifiedEmailChecker[*mocks.MockVerifiedEmailChecker](mod)

			resetPassword := buildResetPassword(t, mod)

			// Create a reset password request
			rp, err := resetPassword.Request(
				context.Background(),
				email,
			)

			savedRp := authFixture.ResetPasswordRequest().ID(rp.ID).PullUpdates(t).Cleanup(t).GetEntity()
			authFixture.Identity().AccountID(rp.AccountID).CleanupAllOfAccount(t)

			t.Log("Given an identity registered via email")
			t.Logf("When I request a reset password for %s", email)
			t.Log("	Then the reset password entity should be saved")
			require.NoError(t, err)
			require.Equal(t, savedRp.ID, rp.ID)
			require.NotEmpty(t, savedRp.Token)
			t.Log("	And email should be sent")
			senderMock.AssertExpectations(t)
			t.Log("	And email should be checked successfully")
			verifierMock.AssertExpectations(t)
			t.Log("	And reset password request should be created for the existent account")
			require.Equal(t, savedRp.AccountID.String(), rp.AccountID.String())
			require.Equal(t, savedRp.AccountID.String(), account.ID.String())
			t.Log("	And identity should be created")
		},
	)
}

func buildResetPassword(t *testing.T, mod *module.Module) *action.ResetPassword {
	var resetPassword *action.ResetPassword
	err := test.Invoke(
		module.BuildFx(mod),
		fx.Populate(&resetPassword),
	)
	require.NoError(t, err)
	return resetPassword
}

func createSenderMockSuccess(t *testing.T, email string, mod *module.Module) *mocks.MockMailSender {
	senderMock := mocks.NewMockMailSender(t)

	senderMock.EXPECT().ResetPasswordEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(
		nil,
	)
	mod.AddProviders(
		func() *mocks.MockMailSender {
			return senderMock
		},
	)
	return senderMock
}

func createVerifiedEmailCheckerMockSuccess(
	t *testing.T,
	email string,
	userID uuid.UUID,
	mod *module.Module,
) *mocks.MockVerifiedEmailChecker {
	checker := mocks.NewMockVerifiedEmailChecker(t)

	checker.EXPECT().FindUserIDByVerifiedEmail(mock.Anything, email).Return(
		userID, nil,
	)
	mod.AddProviders(
		func() *mocks.MockVerifiedEmailChecker {
			return checker
		},
	)
	return checker
}
