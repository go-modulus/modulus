package spec

import (
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/gofrs/uuid"
	"testing"
)

var (
	reset = "\033[0m"
	//nolint:unused
	red = "\033[31m"
	//nolint:unused
	green = "\033[32m"
	cyan  = "\033[36m"
	blue  = "\033[34m"
)

var assertionPrefix = "    "

func When(t *testing.T, msg string) {
	t.Helper()
	t.Log(cyan + "When " + msg + reset)
}

func Context(t *testing.T, msg string) {
	t.Helper()
	t.Log(cyan + msg + reset)
}

func Auth(t *testing.T, ctx context.Context) {
	t.Helper()
	id := auth.GetPerformerID(ctx)
	msg := ""
	if id == uuid.Nil {
		msg = "I am not authenticated"
	} else {
		msg = "I am authenticated as user " + id.String()
	}
	t.Log(cyan + msg + reset)
}

func Given(t testing.TB, descriptions ...string) {
	t.Helper()

	t.Log(cyan + "Given:" + reset)
	for _, desc := range descriptions {
		t.Log(blue + assertionPrefix + desc + reset)
	}
}
