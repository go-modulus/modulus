package graphql

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/auth/graphql"
	graphql2 "github.com/go-modulus/modulus/auth/install/graphql"
	"github.com/go-modulus/modulus/auth/providers/google/action"
)

type Resolver struct {
	register *action.Register
}

func NewResolver(register *action.Register) *Resolver {
	return &Resolver{
		register: register,
	}
}

func (r *Resolver) RegisterViaGoogle(ctx context.Context, input RegisterViaGoogleInput) (graphql.TokenPair, error) {
	url := ""
	if input.RedirectURL != nil {
		url = *input.RedirectURL
	}

	tokens, err := r.register.Execute(
		ctx, action.RegisterInput{
			Code:        input.Code,
			Verifier:    input.Verifier,
			RedirectUrl: url,
			Roles:       []string{graphql2.DefaultUserRole},
			UserInfo:    nil,
		},
	)
	if err != nil {
		return graphql.TokenPair{}, errtrace.Wrap(err)
	}

	auth.SendRefreshToken(ctx, tokens.RefreshToken.Token.String)

	return graphql.TokenPair{
		AccessToken:  tokens.AccessToken.Token.String,
		RefreshToken: tokens.RefreshToken.Token.String,
	}, nil
}
