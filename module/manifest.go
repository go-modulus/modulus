package module

import (
	"encoding/json"
	"io/fs"
)

type ManifestItem struct {
	Name           string `json:"name"`
	Package        string `json:"package"`
	Description    string `json:"description"`
	InstallCommand string `json:"install"`
	Version        string `json:"version"`
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
