package mailtrap_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/vorobeyme/mailtrap-go/mailtrap"
	"testing"
)

func TestSender_Send(t *testing.T) {
	t.Skip("Skip test due to the real email sending")
	t.Run(
		"Send email", func(t *testing.T) {
			err := sender.Send(
				context.Background(),
				"Test subject",
				"<html><body><h1>Test</h1><p>body</p></body></html>",
				[]mailtrap.EmailAddress{
					{
						Email: "yanak1984@gmail.com",
						Name:  "Andrii",
					},
				},
			)

			require.NoError(t, err)
		},
	)
}
