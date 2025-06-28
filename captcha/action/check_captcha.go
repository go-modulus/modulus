package action

import (
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/xinguang/go-recaptcha"
)

var ErrUnknownCaptchaVersion = errsys.New(
	"unknown captcha version",
	"Captcha version is not supported",
)

var ErrCaptchaVerify = errsys.New(
	"captcha verify failed",
	"Captcha verification failed",
)

var ErrRecaptchaServiceNotAvailable = errsys.New(
	"recaptcha service not available",
	"Recaptcha service is not available",
)

var ErrCaptchaTokenIsRequired = errsys.New(
	"captcha token is required",
	"Captcha token is not found. Provide a token to verify",
)

type CaptchaVersion string

const (
	CaptchaVersionV2 CaptchaVersion = "v2"
	CaptchaVersionV3 CaptchaVersion = "v3"
)

type CaptchaToken struct {
	Version CaptchaVersion
	Token   string
	Action  string
}

type RecaptchaConfig struct {
	Enabled     bool    `env:"RECAPTCHA_ENABLED,default=false"`
	SecretV3    string  `env:"RECAPTCHA_V3_SECRET"`
	SecretV2    string  `env:"RECAPTCHA_V2_SECRET"`
	ThresholdV3 float64 `env:"RECAPTCHA_V3_THRESHOLD,default=0.5"`
}

type CheckCaptcha struct {
	config RecaptchaConfig
}

func NewCheckCaptcha(
	config RecaptchaConfig,
) *CheckCaptcha {
	return &CheckCaptcha{
		config: config,
	}
}

func (c *CheckCaptcha) Execute(token *CaptchaToken) error {
	if !c.config.Enabled {
		return nil
	}
	if token == nil {
		return ErrCaptchaTokenIsRequired
	}
	if token.Version == CaptchaVersionV2 {
		return c.checkV2(token.Token)
	}
	if token.Version == CaptchaVersionV3 {
		return c.checkV3(token.Token, token.Action)
	}
	return ErrUnknownCaptchaVersion
}

func (c *CheckCaptcha) checkV2(token string) error {
	captcha, err := recaptcha.NewWithSecert(c.config.SecretV2)
	if err != nil {
		return errsys.WithCause(ErrRecaptchaServiceNotAvailable, err)
	}
	err = captcha.Verify(token)
	if err != nil {
		return errsys.WithCause(ErrCaptchaVerify, err)
	}
	return nil
}

func (c *CheckCaptcha) checkV3(token string, action string) error {
	captcha, err := recaptcha.NewWithSecert(c.config.SecretV3)
	if err != nil {
		return errsys.WithCause(ErrRecaptchaServiceNotAvailable, err)
	}
	err = captcha.VerifyWithOptions(
		token, recaptcha.VerifyOption{
			Action:    action,
			Threshold: c.config.ThresholdV3,
		},
	)
	if err != nil {
		return errsys.WithCause(ErrCaptchaVerify, err)
	}
	return nil
}
