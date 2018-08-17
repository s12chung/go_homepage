package test

import (
	"fmt"
	"sort"
	"strings"
)

type ContextFields map[string]interface{}

type Context struct {
	fields       ContextFields
	fieldsString string
}

func NewContext() *Context {
	return &Context{}
}

func (context *Context) SetFields(fields ContextFields) *Context {
	context.fields = fields
	context.fieldsString = ""
	return context
}

func (context *Context) makeFieldsString() string {
	fieldStrings := make([]string, len(context.fields))
	i := 0
	for k, v := range context.fields {
		fieldStrings[i] = fmt.Sprintf("%v=%v", k, v)
		i += 1
	}
	sort.Strings(fieldStrings)
	return strings.Join(fieldStrings, " ")
}

func (context *Context) FieldsString() string {
	if context.fieldsString == "" {
		context.fieldsString = context.makeFieldsString()
	}
	return context.fieldsString
}

func (context *Context) String(i interface{}) string {
	return context.Stringf("%v", i)
}

func (context *Context) Stringf(format string, args ...interface{}) string {
	return strings.Join([]string{context.FieldsString(), fmt.Sprintf(format, args...)}, " - ")
}

func (context *Context) GotExpString(label string, got, exp interface{}) string {
	return context.Stringf("%v - got: %v, exp: %v", label, got, exp)
}

func (context *Context) DiffString(label string, got, exp, diff interface{}) string {
	return context.Stringf("%v - got: %v, exp: %v, diff: %v", label, got, exp, diff)
}
