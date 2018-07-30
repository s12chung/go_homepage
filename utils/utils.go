package utils

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
)

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

func FilePaths(dirPaths ...string) ([]string, error) {
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
			if fileInfo.IsDir() {
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
	if dataValue.Type().Kind() != reflect.Ptr {
		dataValue = reflect.New(reflect.TypeOf(data))
	}

	field := dataValue.Elem().FieldByName(name)
	if field.IsValid() {
		return field.String()
	} else {
		return ""
	}
}
