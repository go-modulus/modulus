package mailtrap

import (
	"context"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
	_ "golang.org/x/text/message"
)

type ModuleConfig struct {
	FromName     string `env:"MAILTRAP_FROM_NAME, default=TrustyPay"`
	FromEmail    string `env:"MAILTRAP_FROM_EMAIL, default=no-reply@trustypay.com.ua"`
	ApiKey       string `env:"MAILTRAP_API_KEY"`
	SendEmails   bool   `env:"MAILTRAP_SEND_EMAILS, default=false"`
	FrontendHost string `env:"FRONTEND_HOST, required"`
}

func NewModule(config ModuleConfig) fx.Option {
	return fx.Module(
		"mailtrap",
		fx.Provide(
			func() (*ModuleConfig, error) {
				return &config, envconfig.Process(context.Background(), &config)
			},
			NewSender,
			func(sender *RealSender) Sender { return sender },
		),
	)
}
