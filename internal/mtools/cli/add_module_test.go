package cli_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestAddModule_Invoke(t *testing.T) {
	t.Run(
		"update tools.go with new module", func(t *testing.T) {
			if _, err := os.Stat("/tmp/testproj"); os.IsNotExist(err) {
				err = os.Mkdir("/tmp/testproj", 0755)
				if err != nil {
					t.Fatal("Cannot create /tmp/testproj dir", err)
				}
			}
			err := os.Chdir("/tmp/testproj")
			require.NoError(t, err)
			ctx := &cli.Context{
				Context: context.Background(),
			}
			err = addModule.Invoke(ctx)
			t.Log("Given the tools.go file in the root of the project")
			t.Log("When install a new module to a project")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	The new package should be added to the tools.go file")
		},
	)
}
