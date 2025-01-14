package module

import (
	"encoding/json"
	"fmt"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"io/fs"
	"os"
	"strings"
)

var ErrCannotReadEntries = fmt.Errorf("cannot read entries")

type ManifestModule struct {
	Name          string          `json:"name"`
	Package       string          `json:"package"`
	Description   string          `json:"description"`
	Install       InstallManifest `json:"install,omitempty"`
	Version       string          `json:"version"`
	LocalPath     string          `json:"localPath,omitempty"`
	IsLocalModule bool            `json:"isLocalModule,omitempty"`
}

func NewManifestModule(
	module *Module,
	pckg string,
	description string,
	version string,
) ManifestModule {
	deps := make([]string, 0, len(module.dependencies))
	for _, dep := range module.dependencies {
		deps = append(deps, dep.name)
	}
	install := InstallManifest{}
	install.
		AppendEnvVars(module.envVars...).
		AppendDependencies(deps...)

	currentModule := ManifestModule{
		Name:        module.name,
		Package:     pckg,
		Install:     install,
		Version:     version,
		Description: description,
	}

	return currentModule
}

type Entrypoint struct {
	LocalPath string `json:"localPath"`
	Name      string `json:"name"`
}

type Manifest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Modules     []ManifestModule `json:"modules"`
	Entries     []Entrypoint     `json:"entries,omitempty"`
}

func (m *Manifest) ReadFromJSON(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m *Manifest) WriteToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

func (m *Manifest) AddModule(module ManifestModule) {
	m.Modules = append(m.Modules, module)
}

func (m *Manifest) UpdateModule(module ManifestModule) {
	for i, mod := range m.Modules {
		if mod.Package == module.Package {
			m.Modules[i] = module
			return
		}
	}
	m.AddModule(module)
}

func NewFromFs(manifestFs fs.FS, filename string) (*Manifest, error) {
	data, err := fs.ReadFile(manifestFs, filename)
	if err != nil {
		return nil, err
	}
	m := &Manifest{}
	err = m.ReadFromJSON(data)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func LoadLocalManifest(projPath string) (Manifest, error) {
	entries, err := readEntries(projPath)
	if err != nil {
		return Manifest{}, errors.WrapCause(ErrCannotReadEntries, err)
	}
	res := Manifest{
		Modules:     make([]ManifestModule, 0),
		Version:     "1.0.0",
		Name:        "Modulus framework modules manifest",
		Description: "List of installed modules for the Modulus framework",
		Entries:     entries,
	}
	if utils.FileExists(projPath + "/modules.json") {
		projFs := os.DirFS(projPath)
		manifest, err := NewFromFs(projFs, "modules.json")
		if err != nil {
			return res, err
		}
		return *manifest, nil
	}
	return res, nil
}

func readEntries(projPath string) (entries []Entrypoint, err error) {
	folders, err := os.ReadDir(projPath + "/cmd")
	if err != nil {
		return
	}
	entries = make([]Entrypoint, 0, len(folders))
	for _, entry := range folders {
		if entry.IsDir() {
			entryItem := Entrypoint{
				Name: entry.Name(),
			}
			_, err2 := os.Stat(projPath + "/cmd/" + entry.Name() + "/main.go")
			if os.IsNotExist(err2) {
				continue
			}

			if err2 != nil {
				err = err2
				return
			}
			entryItem.LocalPath = "cmd/" + entry.Name() + "/main.go"
			entries = append(entries, entryItem)
		}
	}

	return
}

func (m *Manifest) SaveAsLocalManifest(projPath string) error {
	data, err := m.WriteToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(projPath+"/modules.json", data, 0644)
}

func (m *Manifest) LocalModules() []ManifestModule {
	res := make([]ManifestModule, 0)
	for _, mod := range m.Modules {
		if mod.IsLocalModule {
			res = append(res, mod)
		}
	}
	return res
}

func (m *Manifest) FindLocalModule(moduleName string) (ManifestModule, bool) {
	for _, mod := range m.Modules {
		if mod.IsLocalModule && strings.EqualFold(mod.Name, moduleName) {
			return mod, true
		}
	}
	return ManifestModule{}, false
}

func (m ManifestModule) GetShortPackageName() string {
	return m.Package[strings.LastIndex(m.Package, "/")+1:]
}

func (m ManifestModule) StoragePath(projPath string) string {
	return m.ModulePath(projPath) + "/storage"
}

func (m ManifestModule) ModulePath(projPath string) string {
	return projPath + "/" + m.LocalPath
}

func (m ManifestModule) StoragePackage() string {
	return m.Package + "/storage"
}

func (m ManifestModule) CliPath(projPath string) string {
	return m.ModulePath(projPath) + "/cli"
}

func (m ManifestModule) CliPackage() string {
	return m.Package + "/cli"
}

func (m ManifestModule) ApiPath(projPath string) string {
	return m.ModulePath(projPath) + "/api"
}

func (m ManifestModule) ApiPackage() string {
	return m.Package + "/api"
}
