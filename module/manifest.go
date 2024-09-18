package module

import (
	"encoding/json"
	"io/fs"
)

type ManifestItem struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Description  string `json:"description"`
	Package      string `json:"package"`
	SetupCommand string `json:"setup"`
}
type Manifest struct {
	Modules     []ManifestItem `json:"modules"`
	Version     string         `json:"version"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
}

func (m *Manifest) ReadFromJSON(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m *Manifest) WriteToJSON() ([]byte, error) {
	return json.Marshal(m)
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
