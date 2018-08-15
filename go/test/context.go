package test

import (
	"fmt"
	"sort"
	"strings"
)

type ContextFields map[string]interface{}

type Context struct {
	fieldsString string
}

func NewContext() *Context {
	return &Context{}
}

func (context *Context) SetFields(fields ContextFields) *Context {
	fieldStrings := make([]string, len(fields))
	i := 0
	for k, v := range fields {
		fieldStrings[i] = fmt.Sprintf("%v=%v", k, v)
		i += 1
	}
	sort.Strings(fieldStrings)
	context.fieldsString = strings.Join(fieldStrings, " ")
	return context
}

func (context *Context) String(i interface{}) string {
	return context.Stringf("%v", i)
}

func (context *Context) Stringf(format string, args ...interface{}) string {
	return strings.Join([]string{context.fieldsString, fmt.Sprintf(format, args...)}, " - ")
}

func (context *Context) GotExpString(label string, got, exp interface{}) string {
	return context.Stringf("%v - got: %v, exp: %v", label, got, exp)
}

func (context *Context) DiffString(label string, got, exp, diff interface{}) string {
	return context.Stringf("%v - got: %v, exp: %v, diff: %v", label, got, exp, diff)
}
