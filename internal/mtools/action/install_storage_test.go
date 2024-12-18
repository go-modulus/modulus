package action_test

import (
	"context"
	"fmt"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/go-modulus/modulus/module"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestInstallStorage_Install(t *testing.T) {
	t.Run(
		"Install", func(t *testing.T) {
			projDir := "/tmp/testproj-storage"
			moduleDir := fmt.Sprintf("%s/mypckg", projDir)
			storageDir := fmt.Sprintf("%s/storage", moduleDir)

			err := os.Mkdir(projDir, 0755)
			require.NoError(t, err)
			defer os.RemoveAll(projDir)

			err = os.Mkdir(moduleDir, 0755)
			require.NoError(t, err)

			currentDir, _ := os.Getwd()
			os.Chdir(projDir)

			md := module.ManifestItem{
				Name:           "My package",
				Package:        "mypckg",
				Description:    "",
				InstallCommand: "",
				Version:        "",
				LocalPath:      moduleDir,
			}
			err = installStorage.Install(
				context.Background(), md, action.StorageConfig{
					Schema:             "schema",
					GenerateGraphql:    true,
					GenerateFixture:    true,
					GenerateDataloader: true,
				},
			)

			os.Chdir(currentDir)
			_, errDir1 := os.Stat(storageDir)
			_, errDir2 := os.Stat(storageDir + "/migration")
			_, errDir3 := os.Stat(storageDir + "/query")
			contentDef, errDefFile := os.ReadFile(projDir + "/sqlc.definition.yaml")
			contentTmpl, errTmplFile := os.ReadFile(storageDir + "/sqlc.tmpl.yaml")
			confTmpl, errConfFile := os.ReadFile(storageDir + "/sqlc.yaml")

			t.Log("When install storage to a module")
			t.Log("	The error should be nil")
			require.NoError(t, err)

			t.Log("	The storage directory should be created")
			require.NoError(t, errDir1)
			t.Log("	The migration directory should be created")
			require.NoError(t, errDir2)

			t.Log("	The query directory should be created")
			require.NoError(t, errDir3)

			t.Log("	The sqlc.definition.yaml file should be created")
			require.NoError(t, errDefFile)
			snaps.WithConfig(snaps.Ext(".sqlc.definition.yaml")).
				MatchStandaloneSnapshot(t, string(contentDef))

			t.Log("	The sqlc template file should be created")
			require.NoError(t, errTmplFile)
			snaps.WithConfig(snaps.Ext(".sqlc.tmpl.yaml")).
				MatchStandaloneSnapshot(t, string(contentTmpl))

			t.Log("	The sqlc.yaml file should be created")
			require.NoError(t, errConfFile)
			snaps.WithConfig(snaps.Ext(".sqlc.yaml")).
				MatchStandaloneSnapshot(t, string(confTmpl))
		},
	)
}
