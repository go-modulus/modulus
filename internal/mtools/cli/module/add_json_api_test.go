package module_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func createModuleInTmpDir(projDir string) error {
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	set.String("package", "mypckg", "")
	set.String("name", "mypckg", "")
	set.String("path", "internal", "")
	set.String("proj-path", projDir, "")
	set.Bool("silent", true, "")
	ctx := cli.NewContext(app, set, nil)
	return createModule.Invoke(ctx)
}

func TestAddJsonApi_Invoke(t *testing.T) {
	t.Run(
		"create json api", func(t *testing.T) {
			projDir := "/tmp/testproj"
			rb := initProject(t, projDir, goModFile)
			defer rb()

			err := createModuleInTmpDir(projDir)
			require.NoError(t, err)

			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.String("name", "HelloWorld", "")
			set.String("uri", "/mypckg/hello-world", "")
			set.String("method", "GET", "")
			set.String("module", "mypckg", "")
			set.String("proj-path", projDir, "")
			set.Bool("silent", true, "")
			ctx := cli.NewContext(app, set, nil)

			err = addJsonApi.Invoke(ctx)

			apiDir := fmt.Sprintf("%s/internal/mypckg/api", projDir)
			_, errDir := os.Stat(apiDir)
			handlerContent, errCont1 := os.ReadFile(fmt.Sprintf("%s/hello_world.go", apiDir))

			t.Log("When create a new json api handler in the module")
			t.Log("	The error should be nil")
			require.NoError(t, err)
			t.Log("	The api directory should be created in the module")
			require.NoError(t, errDir)
			t.Log("	The api handler file should be created")
			require.NoError(t, errCont1)
			require.Contains(
				t, string(handlerContent), "type HelloWorld struct {",
			)
		},
	)
}
