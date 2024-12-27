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

type ManifestItem struct {
	Name           string `json:"name"`
	Package        string `json:"package"`
	Description    string `json:"description"`
	InstallCommand string `json:"install"`
	Version        string `json:"version"`
	LocalPath      string `json:"localPath"`
	IsLocalModule  bool   `json:"isLocalModule"`
}

type Entrypoint struct {
	LocalPath string `json:"localPath"`
	Name      string `json:"name"`
}

type Manifest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Modules     []ManifestItem `json:"modules"`
	Entries     []Entrypoint   `json:"entries"`
}

func (m *Manifest) ReadFromJSON(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m *Manifest) WriteToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
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
		Modules:     make([]ManifestItem, 0),
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

func (m ManifestItem) GetShortPackageName() string {
	return m.Package[strings.LastIndex(m.Package, "/")+1:]
}

func (m ManifestItem) StoragePath(projPath string) string {
	return m.ModulePath(projPath) + "/storage"
}

func (m ManifestItem) ModulePath(projPath string) string {
	return projPath + "/" + m.LocalPath
}

func (m ManifestItem) StoragePackage() string {
	return m.Package + "/storage"
}
