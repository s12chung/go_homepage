package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

func SettingsFromFile(path string, settings interface{}, log logrus.FieldLogger) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Infof("%v not found, using defaults...", path)
		return
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", path)
		return
	}

	err = json.Unmarshal(file, settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", path)
		return
	}
}

func MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func WriteFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}

func CleanFilePath(filePath string) string {
	filePath = strings.TrimLeft(filePath, ".")
	return strings.Trim(filePath, "/")
}

func ToSimpleQuery(queryMap map[string]string) string {
	queryArray := make([]string, len(queryMap))
	index := 0
	for key, value := range queryMap {
		queryArray[index] = key + "=" + value
		index += 1
	}
	sort.Strings(queryArray)
	return strings.Join(queryArray, "&")
}

func SliceList(slice []string) string {
	sliceLength := len(slice)
	newSlice := []string{}
	for index, item := range slice {
		newSlice = append(newSlice, item)

		if index == sliceLength-1 {
			continue
		}

		between := ", "
		if index == sliceLength-2 {
			between = " & "
		}
		newSlice = append(newSlice, between)
	}
	return strings.Join(newSlice, "")
}

func FilePaths(suffix string, dirPaths ...string) ([]string, error) {
	var filePaths []string

	for _, dirPath := range dirPaths {
		_, err := os.Stat(dirPath)
		if err != nil {
			return nil, err
		}
		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}

		for _, fileInfo := range files {
			if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), suffix) {
				continue
			}
			filePaths = append(filePaths, path.Join(dirPath, fileInfo.Name()))
		}
	}
	return filePaths, nil
}

func GetStringField(data interface{}, name string) string {
	if data == nil {
		return ""
	}

	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	field := dataValue.FieldByName(name)
	if field.IsValid() && field.Type().Kind() == reflect.String {
		return field.String()
	}
	return ""
}
