package graphql

import captchaAction "github.com/go-modulus/modulus/captcha/action"

type RegisterViaEmailInput struct {
	Email    string                      `json:"email"`
	Password string                      `json:"password"`
	UserInfo map[string]interface{}      `json:"userInfo"`
	Captcha  *captchaAction.CaptchaToken `json:"captcha"`
}

type LoginViaEmailInput struct {
	Email    string                      `json:"email"`
	Password string                      `json:"password"`
	Captcha  *captchaAction.CaptchaToken `json:"captcha"`
}
