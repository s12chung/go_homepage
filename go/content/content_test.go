package content

import (
	"testing"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/test"
	"sort"
)

func defaultContent() (*Content, logrus.FieldLogger, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewContent("", DefaultSettings(), log), log, hook
}

var handler = func(ctx router.Context) error {
	return nil
}

type routeOne struct {
}

func (one *routeOne) SetRoutes(r router.Router, tracker *app.Tracker) error {
	r.GetRootHTML(handler)
	r.GetHTML("/about", handler)
	r.GetHTML("/posts", handler)
	r.Get("/robots.txt", handler)

	tracker.AddDependentURL(router.RootURL)
	tracker.AddDependentURL("/posts")
	return nil
}
func (one *routeOne) WildcardURLs() ([]string, error) {
	return []string{"one", "two", "three"}, nil
}

type routeTwo struct {
}

func (two *routeTwo) SetRoutes(r router.Router, tracker *app.Tracker) error {
	r.GetHTML("/something", handler)
	r.Get("/posts.atom", handler)
	tracker.AddDependentURL("/something")
	return nil
}
func (two *routeTwo) WildcardURLs() ([]string, error) {
	return []string{"two", "three", "four", "five"}, nil
}

type routeThree struct {
}

func (three *routeThree) SetRoutes(r router.Router, tracker *app.Tracker) error {
	r.GetHTML("/about", handler)
	return nil
}

func TestContent_SetRoutes(t *testing.T) {
	testCases := []struct {
		routes        []Route
		urls          []string
		dependentURLs []string
	}{
		{[]Route{}, []string{}, []string{}},
		{[]Route{&routeOne{}}, []string{"/", "/about", "/posts", "/robots.txt"}, []string{"/posts", router.RootURL}},
		{[]Route{&routeTwo{}}, []string{"/something", "/posts.atom"}, []string{"/something"}},
		{[]Route{&routeThree{}}, []string{"/about"}, []string{}},
		{[]Route{&routeOne{}, &routeTwo{}}, []string{"/", "/about", "/posts", "/robots.txt", "/something", "/posts.atom"}, []string{"/posts", router.RootURL, "/something"}},
		{[]Route{&routeTwo{}, &routeThree{}}, []string{"/something", "/posts.atom", "/about"}, []string{"/something"}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":  testCaseIndex,
			"routes": tc.routes,
		})

		content, log, _ := defaultContent()
		content.routes = tc.routes
		r := router.NewGenerateRouter(log)
		tracker := app.NewTracker(func() []string {
			return nil
		})

		err := content.SetRoutes(r, tracker)
		if err != nil {
			t.Error(context.String(err))
		}

		got := r.URLs()
		sort.Strings(got)
		sort.Strings(tc.urls)
		if !cmp.Equal(got, tc.urls) {
			t.Error(context.DiffString("r.URLs()", got, tc.urls, cmp.Diff(got, tc.urls)))
		}

		got = tracker.DependentURLs()
		sort.Strings(got)
		sort.Strings(tc.dependentURLs)
		if !cmp.Equal(got, tc.dependentURLs) {
			t.Error(context.DiffString("tracker.DependentURLs()", got, tc.urls, cmp.Diff(got, tc.dependentURLs)))
		}
	}
}
