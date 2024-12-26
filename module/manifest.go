package module

import (
	"encoding/json"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"io/fs"
	"os"
	"strings"
)

type ManifestItem struct {
	Name           string `json:"name"`
	Package        string `json:"package"`
	Description    string `json:"description"`
	InstallCommand string `json:"install"`
	Version        string `json:"version"`
	LocalPath      string `json:"localPath"`
	IsLocalModule  bool   `json:"isLocalModule"`
}
type Manifest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Modules     []ManifestItem `json:"modules"`
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
	res := Manifest{
		Modules:     make([]ManifestItem, 0),
		Version:     "1.0.0",
		Name:        "Modulus framework modules manifest",
		Description: "List of installed modules for the Modulus framework",
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
