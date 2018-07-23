package utils

import (
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
