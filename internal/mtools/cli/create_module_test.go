package cli_test

import (
	"flag"
	"fmt"
	"github.com/go-modulus/modulus/module"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestCreateModule_Invoke(t *testing.T) {
	t.Run(
		"create module", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir)
			defer rb()

			err := os.Chdir(projDir)
			require.NoError(t, err)
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.String("package", "mypckg", "")
			set.String("path", "internal/mypckg", "")
			ctx := cli.NewContext(app, set, nil)
			err = createModule.Invoke(ctx)

			moduleDir := fmt.Sprintf("%s/internal/mypckg", projDir)
			_, errDir := os.Stat(moduleDir)

			localManifest, errCont := module.LoadLocalManifest()
			moduleContent, errCont1 := os.ReadFile(fmt.Sprintf("%s/module.go", moduleDir))

			t.Log("When create a new module to a project")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	The module directory should be created")
			require.NoError(t, errDir)
			t.Log("	The new module should be added to the local manifest")
			require.NoError(t, errCont)
			require.Contains(
				t, localManifest.Modules, module.ManifestItem{
					Name:           "mypckg",
					Package:        "testproj/internal/mypckg",
					Description:    "",
					InstallCommand: "",
					Version:        "",
					LocalPath:      "internal/mypckg",
				},
			)
			t.Log("The module file should be created")
			require.NoError(t, errCont1)
			require.Contains(
				t, string(moduleContent), "package mypckg",
			)
		},
	)
}
