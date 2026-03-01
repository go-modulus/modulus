package module

import (
	"fmt"
	"strings"
)

var ErrCannotReadEntries = fmt.Errorf("cannot read entries")

type Manifesto struct {
	Name          string                `json:"name"`
	Package       string                `json:"package"`
	Description   string                `json:"description"`
	Install       InstallationManifesto `json:"install,omitempty"`
	Version       string                `json:"version"`
	LocalPath     string                `json:"localPath,omitempty"`
	IsLocalModule bool                  `json:"isLocalModule,omitempty"`
}

func NewManifesto(
	module *Module,
	pkg string,
	description string,
	version string,
) Manifesto {
	deps := make([]string, 0, len(module.dependencies))
	for _, dep := range module.dependencies {
		deps = append(deps, dep.name)
	}
	install := InstallationManifesto{}
	if len(module.envVars) > 0 {
		install.
			AppendEnvVars(module.envVars...)
	}
	if len(deps) > 0 {
		install.
			AppendDependencies(deps...)
	}

	currentModule := Manifesto{
		Name:        module.name,
		Package:     pkg,
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

//type Manifest struct {
//	Name        string           `json:"name"`
//	Description string           `json:"description"`
//	Version     string       `json:"version"`
//	Modules     []Manifesto  `json:"modules"`
//	Entries     []Entrypoint `json:"entries,omitempty"`
//}
//
//func (m *Manifest) ReadFromJSON(data []byte) error {
//	return json.Unmarshal(data, &m)
//}
//
//func (m *Manifest) WriteToJSON() ([]byte, error) {
//	return json.MarshalIndent(m, "", "  ")
//}
//
//func (m *Manifest) AddModule(module Manifesto) {
//	m.Modules = append(m.Modules, module)
//}
//
//func (m *Manifest) UpdateModule(module Manifesto) {
//	for i, mod := range m.Modules {
//		if mod.Package == module.Package {
//			m.Modules[i] = module
//			return
//		}
//	}
//	m.AddModule(module)
//}
//
//func NewFromFs(manifestFs fs.FS, filename string) (*Manifest, error) {
//	data, err := fs.ReadFile(manifestFs, filename)
//	if err != nil {
//		return nil, err
//	}
//	m := &Manifest{}
//	err = m.ReadFromJSON(data)
//	if err != nil {
//		return nil, err
//	}
//	return m, nil
//}
//
////func (m *Manifest) NewFromFullPath(fullFileName string) (*Manifest, error) {
////	if utils.FileExists(fullFileName) {
////		path := strings.Split(fullFileName, "/")
////		projPath := strings.Join(path[:len(path)-1], "/")
////		fileName := path[len(path)-1]
////		projFs := os.DirFS(projPath)
////		manifest, err := NewFromFs(projFs, fileName)
////		if err != nil {
////			return nil, err
////		}
////		return manifest, nil
////	}
////}
//
//func LoadLocalManifest(projPath string) (Manifest, error) {
//	entries, err := readEntries(projPath)
//	if err != nil {
//		return Manifest{}, errors.WithCause(ErrCannotReadEntries, err)
//	}
//	res := Manifest{
//		Modules:     make([]Manifesto, 0),
//		Version:     "1.0.0",
//		Name:        "Modulus framework modules manifest",
//		Description: "List of installed modules for the Modulus framework",
//		Entries:     entries,
//	}
//	if fileExists(projPath + "/modules.json") {
//		projFs := os.DirFS(projPath)
//		manifest, err := NewFromFs(projFs, "modules.json")
//		if err != nil {
//			return res, err
//		}
//		return *manifest, nil
//	}
//	return res, nil
//}

//func fileExists(filename string) bool {
//	info, err := os.Stat(filename)
//	if os.IsNotExist(err) {
//		return false
//	}
//	return !info.IsDir()
//}
//
//func readEntries(projPath string) (entries []Entrypoint, err error) {
//	folders, err := os.ReadDir(projPath + "/cmd")
//	if err != nil {
//		return
//	}
//	entries = make([]Entrypoint, 0, len(folders))
//	for _, entry := range folders {
//		if entry.IsDir() {
//			entryItem := Entrypoint{
//				Name: entry.Name(),
//			}
//			_, err2 := os.Stat(projPath + "/cmd/" + entry.Name() + "/main.go")
//			if os.IsNotExist(err2) {
//				continue
//			}
//
//			if err2 != nil {
//				err = err2
//				return
//			}
//			entryItem.LocalPath = "cmd/" + entry.Name() + "/main.go"
//			entries = append(entries, entryItem)
//		}
//	}
//
//	return
//}

func (m Manifesto) GetShortPackageName() string {
	return m.Package[strings.LastIndex(m.Package, "/")+1:]
}

func (m Manifesto) StoragePath(projPath string) string {
	return m.ModulePath(projPath) + "/storage"
}

func (m Manifesto) ModulePath(projPath string) string {
	if projPath == "" {
		return m.LocalPath
	}
	return projPath + "/" + m.LocalPath
}

func (m Manifesto) StoragePackage() string {
	return m.Package + "/storage"
}

func (m Manifesto) CliPath(projPath string) string {
	return m.ModulePath(projPath) + "/cli"
}

func (m Manifesto) CliPackage() string {
	return m.Package + "/cli"
}

func (m Manifesto) ApiPath(projPath string) string {
	return m.ModulePath(projPath) + "/api"
}

func (m Manifesto) ApiPackage() string {
	return m.Package + "/api"
}
