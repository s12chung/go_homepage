package webpack

import (
	"encoding/json"
	"io/ioutil"
)

func ReadManifest(manifestPath string) (map[string]string, error) {
	manifestMap := map[string]string{}

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &manifestMap)
	return manifestMap, nil
}
