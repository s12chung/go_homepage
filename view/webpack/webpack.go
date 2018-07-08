package webpack

import (
	"encoding/json"
	"io/ioutil"
)

func ReadManifest(manifestPath string) (map[string]string, error) {
	manifestMap := make(map[string]string)

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return manifestMap, err
	}
	json.Unmarshal(data, &manifestMap)
	return manifestMap, nil
}
