package routes

import (
	"testing"

	"github.com/s12chung/gostatic/go/test"
)

func TestTemplateName(t *testing.T) {
	testCases := []struct {
		tmplName string
		exp      string
	}{
		{"abc", "abc"},
		{"abc.html", "abc"},
		{"/abc", "abc"},
		{"/abc.html", "abc"},
		{"wee/abc", "abc"},
		{"/wee/abc", "abc"},
		{"/wee/abc.html", "abc"},
		{"wee/abc.html", "abc"},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":    testCaseIndex,
			"tmplName": tc.tmplName,
		})

		got := templateName(tc.tmplName)
		if got != tc.exp {
			t.Error(context.GotExpString("Result", got, tc.exp))
		}
	}
}
