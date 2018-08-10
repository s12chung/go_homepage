package html

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/s12chung/go_homepage/go/lib/utils"
)

func defaultTemplateFuncs() template.FuncMap {
	scratch := NewScratch()

	return template.FuncMap{
		"scratch": func() *Scratch { return scratch },

		"htmlSafe": htmlSafe,

		"sliceMake": sliceMake,
		"dictMake":  dictMake,
		"sequence":  sequence,

		"dateFormat": dateFormat,
		"now":        time.Now,

		"sliceList": utils.SliceList,
		"toLower":   strings.ToLower,

		"add":      add,
		"subtract": subtract,
		"percent":  percent,
	}
}

// inspired by: https://gohugo.io/functions/scratch
type Scratch struct {
	M map[string]interface{}
}

func NewScratch() *Scratch {
	return &Scratch{
		map[string]interface{}{},
	}
}

func (s *Scratch) Set(key string, value interface{}) string {
	s.M[key] = value
	return ""
}

func (s *Scratch) Append(key string, value interface{}) string {
	if !s.HasKey(key) {
		s.M[key] = []interface{}{value}
	} else {
		s.M[key] = append(s.M[key].([]interface{}), value)
	}
	return ""
}

func (s *Scratch) Get(key string) interface{} {
	return s.M[key]
}

func (s *Scratch) HasKey(key string) bool {
	_, hasKey := s.M[key]
	return hasKey
}

func (s *Scratch) Delete(key string) {
	delete(s.M, key)
}

func htmlSafe(s string) template.HTML {
	return template.HTML(s)
}

func sliceMake(args ...interface{}) []interface{} {
	return args
}

func dictMake(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid Dict call, need to match keys with values")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func sequence(n int) []int {
	seq := make([]int, n)
	for i := range seq {
		seq[i] = i
	}
	return seq
}

func dateFormat(date time.Time) string {
	return date.Format("January 2, 2006")
}

func add(a, b int) int {
	return a + b
}
func subtract(a, b int) int {
	return a - b
}

func percent(a, b int) float32 {
	return (float32(a) / float32(b)) * 100
}
