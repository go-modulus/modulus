package action

import (
	"braces.dev/errtrace"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
)

const IdentityTypeGoogle = "google"

var ErrFailedToGetUserInfo = errsys.New(
	"failed to get user info",
	"There is an issue to get user info from Google. Please try again later",
)
var ErrFailedToExchangeToken = errsys.New(
	"failed to exchange token",
	"There is an issue to exchange Google token. Please try again later or contact support",
)
var ErrInvalidIdentity = errsys.New(
	"invalid identity",
	"The Google registration is broken. We cannot authenticate you at the moment. Please try again later or use another type of registration.",
)
var ErrAccountIsBlocked = errsys.New("account is blocked", "Your account is blocked. Please contact support.")

type RegisterInput struct {
	Code        string                 `json:"code"`
	Verifier    string                 `json:"verifier"`
	RedirectUrl string                 `json:"redirectUrl"`
	Roles       []string               `json:"roles"`
	UserInfo    map[string]interface{} `json:"userInfo"`
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type GoogleConfig struct {
	RedirectUrl  string   `env:"AUTH_GOOGLE_REDIRECT_URL, default=http://localhost:8001/auth/google/callback"`
	ClientID     string   `env:"AUTH_GOOGLE_CLIENT_ID"`
	ClientSecret string   `env:"AUTH_GOOGLE_SECRET"`
	Scopes       []string `env:"AUTH_GOOGLE_SCOPES, default=openid,email,profile"`
}

type Register struct {
	config             GoogleConfig
	plainTokenAuth     *auth.PlainTokenAuthenticator
	identityRepository repository.IdentityRepository
	accountRepository  repository.AccountRepository
	userCreator        UserCreator
}

func NewRegister(
	config GoogleConfig,
	plainTokenAuth *auth.PlainTokenAuthenticator,
	identityRepository repository.IdentityRepository,
	accountRepository repository.AccountRepository,
	creator UserCreator,
) *Register {
	return &Register{
		config:             config,
		plainTokenAuth:     plainTokenAuth,
		identityRepository: identityRepository,
		accountRepository:  accountRepository,
		userCreator:        creator,
	}
}

func (r *Register) Execute(ctx context.Context, input RegisterInput) (auth.TokenPair, error) {
	oauthConfig := &oauth2.Config{
		RedirectURL:  r.config.RedirectUrl,
		ClientID:     r.config.ClientID,
		ClientSecret: r.config.ClientSecret,
		Scopes:       r.config.Scopes,
		Endpoint:     google.Endpoint,
	}
	if input.RedirectUrl != "" {
		oauthConfig.RedirectURL = input.RedirectUrl
	}

	token, err := oauthConfig.Exchange(
		ctx,
		input.Code,
		oauth2.SetAuthURLParam("code_verifier", input.Verifier),
	)
	if err != nil {
		var gErr *oauth2.RetrieveError
		if errors.As(err, &gErr) {
			return auth.TokenPair{}, errsys.WithCause(
				ErrFailedToExchangeToken,
				errsys.New(gErr.ErrorCode, gErr.ErrorDescription),
			)
		}
		return auth.TokenPair{}, errsys.WithCause(ErrFailedToExchangeToken, err)
	}

	googleUser, err := r.getGoogleUserInfo(ctx, oauthConfig, token)
	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(errsys.WithCause(ErrFailedToGetUserInfo, err))
	}

	pair, err := r.auth(ctx, googleUser.ID)
	if err == nil {
		return pair, nil
	}

	// try to register the user
	if !errors.Is(err, ErrInvalidIdentity) {
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	userId := uuid.Must(uuid.NewV6())
	info := r.addGoogleUserToInfo(input.UserInfo, googleUser)
	existingUser, err := r.userCreator.CreateUserFromGoogle(
		ctx, User{
			ID:         userId,
			Email:      googleUser.Email,
			GoogleUser: googleUser,
			UserInfo:   info,
		},
	)
	err = r.processUserCreationError(ctx, err, googleUser, existingUser)
	if err != nil {
		return auth.TokenPair{}, err
	}

	if existingUser.ID != uuid.Nil {
		userId = existingUser.ID
	}

	// Register new account with the Google id as identity
	_, errAccount := r.accountRepository.Create(ctx, userId, input.Roles, input.UserInfo)
	if errAccount != nil {
		if !errors.Is(errAccount, repository.ErrAccountExists) {
			return auth.TokenPair{}, errtrace.Wrap(errAccount)
		}
	}
	_, err = r.identityRepository.Create(
		ctx,
		googleUser.ID,
		userId,
		IdentityTypeGoogle,
		info,
	)
	if err != nil {
		if errAccount == nil {
			_ = r.accountRepository.RemoveAccount(ctx, existingUser.ID)
		}

		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	pair, err = r.auth(ctx, googleUser.ID)
	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	return pair, nil
}

func (r *Register) addGoogleUserToInfo(userInfo map[string]interface{}, googleUser GoogleUser) map[string]interface{} {
	if userInfo == nil {
		userInfo = make(map[string]interface{})
	}
	userInfo["google"] = googleUser
	return userInfo

}

func (r *Register) processUserCreationError(
	ctx context.Context,
	err error,
	googleUser GoogleUser,
	existingUser User,
) error {
	if err == nil {
		return nil
	}

	// If the user already exists with such an email, and the contract of interface is met,
	// try to help the user to log in using another type of authentication.
	if errors.Is(err, ErrUserAlreadyExists) && existingUser.ID != uuid.Nil {
		// if email is verified, then we can link a new auth method to the existing user
		if googleUser.VerifiedEmail {
			return nil
		}
		// if email is not verified, then we should send an error to the user
		identities, err2 := r.identityRepository.GetByAccountID(ctx, existingUser.ID)
		if err2 != nil {
			return errtrace.Wrap(err2)
		}
		if len(identities) > 0 {
			return errors.WithHint(
				ErrUserAlreadyExists,
				fmt.Sprintf(
					"Please log in using another type of authentication. You have registered using %s.",
					identities[0].Type,
				),
			)
		} else {
			return errors.WithMeta(
				errors.WithHint(
					ErrUserAlreadyExists,
					"Your email of the Google account is not verified. Please verify your email first and try again.",
				),
				"user.email", existingUser.Email,
			)
		}
	}
	return errtrace.Wrap(err)
}

func (r *Register) auth(ctx context.Context, googleUserId string) (auth.TokenPair, error) {
	identityObj, err := r.identityRepository.Get(ctx, googleUserId)
	if err != nil {
		if errors.Is(err, repository.ErrIdentityNotFound) {
			return auth.TokenPair{}, errtrace.Wrap(ErrInvalidIdentity)
		}
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	if identityObj.IsBlocked() {
		return auth.TokenPair{}, errtrace.Wrap(ErrAccountIsBlocked)
	}

	// Issue a new pair of access and refresh tokens.
	pair, err := r.plainTokenAuth.IssueTokens(ctx, identityObj.ID, nil)
	if err != nil {
		return auth.TokenPair{}, errtrace.Wrap(err)
	}

	return pair, nil
}

func (r *Register) getGoogleUserInfo(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (
	GoogleUser,
	error,
) {
	client := oauthConfig.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return GoogleUser{}, errtrace.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GoogleUser{}, errtrace.Wrap(fmt.Errorf("unexpected status: %s", resp.Status))
	}

	var userInfo GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return GoogleUser{}, errtrace.Wrap(fmt.Errorf("failed to decode user info: %w", err))
	}

	return userInfo, nil
}
