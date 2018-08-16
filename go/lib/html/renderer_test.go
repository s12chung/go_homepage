package html

import (
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/test"
	"strconv"
)

var updateFixturesPtr = test.UpdateFixtureFlag()

func defaultRenderer() (*Renderer, *logTest.Hook) {
	settings := &Settings{
		test.FixturePath,
		"The Website Title",
		path.Join(test.FixturePath, "markdowns"),
	}
	log, hook := logTest.NewNullLogger()
	return NewRenderer(path.Join(test.FixturePath, "generated"), settings, log), hook
}

func TestRenderer_AssetsUrl(t *testing.T) {
	renderer, _ := defaultRenderer()
	got := renderer.AssetsUrl()
	exp := "/assets/"
	if got != exp {
		test.AssertBasic(t, "Result", got, exp)
	}
}

func TestRenderer_Webpack(t *testing.T) {
	renderer, _ := defaultRenderer()
	got := renderer.Webpack()
	if got != renderer.w {
		test.AssertBasic(t, "Result", got, renderer.w)
	}
}

func TestRenderer_GeneratedAssetsPath(t *testing.T) {
	renderer, _ := defaultRenderer()
	got := renderer.GeneratedAssetsPath()
	if got != renderer.w.GeneratedAssetsPath() {
		test.AssertBasic(t, "Result", got, renderer.w.GeneratedAssetsPath())
	}
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
		{"helpers", "", map[string]string{"Html": `<span>html_data</span>`}},
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
