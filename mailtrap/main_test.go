package mailtrap_test

import (
	"github.com/go-modulus/modulus/mailtrap"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"testing"
)

var sender *mailtrap.RealSender

func TestMain(m *testing.M) {
	test.TestMain(
		m,
		fx.Populate(&sender),
	)
}
