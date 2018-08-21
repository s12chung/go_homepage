package webpack

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

const manifestPath = "manifest.json"

type Manifest struct {
	generatedPath string
	assetsFolder  string
	manifestMap   map[string]string
	log           logrus.FieldLogger
}

func NewManifest(generatedPath, assetsFolder string, log logrus.FieldLogger) *Manifest {
	return &Manifest{generatedPath, assetsFolder, map[string]string{}, log}
}

func (w *Manifest) ManifestUrl(key string) string {
	return w.assetsFolder + "/" + w.manifestValue(key)
}

func (w *Manifest) manifestValue(key string) string {
	if len(w.manifestMap) == 0 {
		err := w.readManifest()
		if err != nil {
			w.log.Errorf("readManifest error: %v", err)
			return key
		}
	}
	value := w.manifestMap[key]
	if value == "" {
		w.log.Errorf("webpack manifestValue not found for key: %v", key)
		return key
	}
	return value
}

func (w *Manifest) readManifest() error {
	bytes, err := ioutil.ReadFile(filepath.Join(w.generatedPath, w.assetsFolder, manifestPath))
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &w.manifestMap)
}
