package webpack

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"regexp"
)

const manifestFilename = "manifest.json"

type Helper struct {
	remoteAssetsPath string
	assetMap         map[string]string
}

func NewHelper(assetsPath string) (*Helper, error) {
	assetMap, err := readManifest(path.Join(assetsPath, manifestFilename))
	if err != nil {
		return nil, err
	}

	return &Helper{
		regexp.MustCompile("\\A.*/").ReplaceAllString(assetsPath, "/"),
		assetMap,
	}, nil
}

func (h *Helper) Url(asset string) string {
	return h.remoteAssetsPath + "/" + asset
}

func readManifest(manifestPath string) (map[string]string, error) {
	assetMap := make(map[string]string)

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return assetMap, err
	}
	json.Unmarshal(data, &assetMap)
	return assetMap, nil
}
