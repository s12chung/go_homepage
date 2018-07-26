package webpack

import (
	"encoding/json"
	"io/ioutil"
)

func ReadManifest(manifestPath string) (map[string]string, error) {
	manifestMap := map[string]string{}

	bytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(bytes, &manifestMap)
	return manifestMap, nil
}
