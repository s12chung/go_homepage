package html

import (
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/lib/markdown"
	"github.com/s12chung/go_homepage/go/lib/webpack"
	"github.com/s12chung/go_homepage/go/test"
)

var updateFixturesPtr = test.UpdateFixtureFlag()

func defaultRenderer() (*Renderer, *logTest.Hook) {
	settings := DefaultSettings()
	settings.TemplatePath = test.FixturePath
	log, hook := logTest.NewNullLogger()

	w := webpack.NewWebpack(path.Join(test.FixturePath, "generated"), webpack.DefaultSettings(), log)
	md := markdown.NewMarkdown(&markdown.Settings{path.Join(test.FixturePath, "markdowns")}, log)
	return NewRenderer(settings, []Plugin{w, md}, log), hook
}

func TestRenderer_Render(t *testing.T) {
	renderer, hook := defaultRenderer()

	testCases := []struct {
		name         string
		defaultTitle string
		data         interface{}
	}{
		{"title", "", nil},
		{"title", "The Default", nil},
		{"title", "The Default", struct{ Title string }{"The Given"}},
		{"title", "", struct{ Title string }{"The Given"}},
		{"renderer_funcs", "", nil},
		{"helpers", "", map[string]interface{}{"Html": `<span>html_data</span>`, "Date": test.Time(1)}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"name":         tc.name,
			"defaultTitle": tc.defaultTitle,
			"data":         tc.data,
		})

		renderered, err := renderer.Render(tc.name, tc.defaultTitle, tc.data)
		if err != nil {
			test.PrintLogEntries(t, hook)
			t.Error(context.String(err))
		}

		got := strings.TrimSpace(string(renderered))

		fixtureName := tc.name + ".html"
		if tc.name == "title" {
			fixtureName = tc.name + strconv.Itoa(testCaseIndex) + ".html"
		}
		if *updateFixturesPtr {
			test.WriteFixture(t, fixtureName, []byte(got))
			continue
		}

		exp := strings.TrimSpace(string(test.ReadFixture(t, fixtureName)))
		if got != exp {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}
