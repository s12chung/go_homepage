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
