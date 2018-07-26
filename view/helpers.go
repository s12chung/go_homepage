package view

import (
	"html/template"
	"strings"
	"time"

	"github.com/s12chung/go_homepage/utils"
)

func defaultTemplateFuncs() template.FuncMap {
	scratch := NewScratch()

	return template.FuncMap{
		"Scratch": func() *Scratch { return scratch },

		"HtmlSafe": htmlSafe,

		"SliceMake": sliceMake,
		"Sequence":  sequence,

		"DateFormat": dateFormat,
		"SliceList":  utils.SliceList,
		"ToLower":    strings.ToLower,

		"Add":      add,
		"Subtract": subtract,
		"Percent":  percent,
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
