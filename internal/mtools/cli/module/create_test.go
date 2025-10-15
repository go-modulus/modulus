package module_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/go-modulus/modulus/module"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCreateModule_Invoke(t *testing.T) {
	t.Run(
		"create module", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir, goModFile)
			defer rb()

			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.String("package", "mypckg", "")
			set.String("path", "internal", "")
			set.String("proj-path", projDir, "")
			set.Bool("silent", true, "")
			ctx := cli.NewContext(app, set, nil)
			err := createModule.Invoke(ctx)

			moduleDir := fmt.Sprintf("%s/internal/mypckg", projDir)
			_, errDir := os.Stat(moduleDir)

			storageDir := fmt.Sprintf("%s/internal/mypckg/storage", projDir)
			_, errStorageDir := os.Stat(storageDir)

			localManifest, errCont := module.LoadLocalManifest(projDir)
			moduleContent, errCont1 := os.ReadFile(fmt.Sprintf("%s/module.go", moduleDir))
			tmplYaml, errCont2 := os.ReadFile(fmt.Sprintf("%s/sqlc.tmpl.yaml", storageDir))
			defStorageYaml, errCont3 := os.ReadFile(fmt.Sprintf("%s/sqlc.definition.yaml", projDir))

			t.Log("When create a new module to a project")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	The module directory should be created")
			require.NoError(t, errDir)
			t.Log("	The new module should be added to the local manifest")
			require.NoError(t, errCont)
			require.Contains(
				t, localManifest.Modules, module.ManifestModule{
					Name:          "mypckg",
					Package:       "testproj/internal/mypckg",
					Description:   "",
					Version:       "",
					LocalPath:     "internal/mypckg",
					IsLocalModule: true,
				},
			)
			t.Log("	The module file should be created")
			require.NoError(t, errCont1)
			require.Contains(
				t, string(moduleContent), "package mypckg",
			)
			t.Log("	Storage feature should be installed")
			t.Log("		The storage directory should be created")
			require.NoError(t, errStorageDir)
			t.Log("		The sqlc definition file should be created")
			require.NoError(t, errCont3)
			require.Contains(
				t, string(defStorageYaml), "default-overrides: &default-overrides",
			)
			t.Log("		The sqlc template file should be created")
			require.NoError(t, errCont2)
			require.Contains(
				t, string(tmplYaml), "sqlc-tmpl",
			)
		},
	)
}
