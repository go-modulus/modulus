package files_test

import (
	"fmt"
	"github.com/go-modulus/modulus/internal/mtools/files"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
	"os"
	"testing"
)

var fileContent = `//go:build tools
// +build tools

package tools

import _ "github.com/vektra/mockery/v2"
import _ "github.com/rakyll/gotest"

`

func TestAddPackageToGoFile(t *testing.T) {
	t.Run(
		"Add a new package to the go file if import is not exist", func(t *testing.T) {

			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(fileContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}

			err = files.AddImportToGoFile("github.com/stretchr/testify", "_", fn)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a go file")
			t.Log("When add a new package to the go file")
			t.Log("	The new package should be added to the go file")
			require.Contains(t, string(fc), "import _ \"github.com/stretchr/testify\"")
		},
	)

	t.Run(
		"Do nothing if package is exist", func(t *testing.T) {

			fn := fmt.Sprintf("/tmp/%s.go", randstr.String(10))
			err := os.WriteFile(fn, []byte(fileContent), 0644)
			defer os.Remove(fn)
			if err != nil {
				t.Fatal("Cannot create "+fn+" file", err)
			}

			err = files.AddImportToGoFile("github.com/rakyll/gotest", "a", fn)
			require.NoError(t, err)
			fc, err := os.ReadFile(fn)
			require.NoError(t, err)

			t.Log("Given a go file")
			t.Log("When add the existent package to the go file")
			t.Log("	The old package should not be changed")
			require.Contains(t, string(fc), "import _ \"github.com/rakyll/gotest\"")
			t.Log("	The new package should not be added")
			require.NotContains(t, string(fc), "import a \"github.com/rakyll/gotest\"")
		},
	)
}

func TestAddImportToTools(t *testing.T) {
	t.Run(
		"Create tools.go if not exists", func(t *testing.T) {
			dir := fmt.Sprintf("/tmp/%s", randstr.String(10))
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.Mkdir(dir, 0755)
				if err != nil {
					t.Fatal("Cannot create /tmp/testproj dir", err)
				}
			}
			defer os.Remove("tools.go")
			defer os.Remove(dir)

			err := os.Chdir(dir)
			require.NoError(t, err)

			err = files.AddImportToTools("github.com/stretchr/testify")
			require.NoError(t, err)

			fc, err := os.ReadFile("tools.go")
			require.NoError(t, err)
			t.Log("Given a go file")
			t.Log("When add a new package to the go file")
			t.Log("	The new package should be added to the tools.go file")
			require.Contains(t, string(fc), "import _ \"github.com/stretchr/testify\"")
			t.Log("The tools.go file should be created with package tools")
			require.Contains(t, string(fc), "package tools")
		},
	)
}
