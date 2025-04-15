package vite

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

type viteManifestData = map[string]ManifestItem

type Manifest struct {
	data viteManifestData
}

type ManifestItem struct {
	File    string   `json:"file"`
	Name    string   `json:"name"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	Css     []string `json:"css"`
}

func FromDistFS(frontend fs.FS) (manifest *Manifest, err error) {
	f, err := frontend.Open(".vite/manifest.json")
	if err != nil {
		return nil, err
	}
	defer func() {
		cErr := f.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}()

	return FromJSON(f)
}

func FromJSON(r io.Reader) (*Manifest, error) {
	var data viteManifestData

	err := json.NewDecoder(r).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &Manifest{
		data,
	}, nil
}

func (m *Manifest) Version() (string, error) {
	hash := sha1.New()
	err := json.NewEncoder(hash).Encode(m.data)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (m *Manifest) GetItem(name string) (ManifestItem, error) {
	entry, ok := m.data[name]
	if !ok {
		return ManifestItem{}, errors.New("item not present in manifest")
	}
	return entry, nil
}
