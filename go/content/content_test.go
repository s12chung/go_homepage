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

func (one *routeOne) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetRootHTML(handler)
	r.GetWildcardHTML(handler)
	r.GetHTML("/about", handler)
	r.GetHTML("/posts", handler)
	r.Get("/robots.txt", handler)

	tracker.AddDependentUrl(router.RootUrlPattern)
	tracker.AddDependentUrl("/posts")
}
func (one *routeOne) WildcardUrls() ([]string, error) {
	return []string{"one", "two", "three"}, nil
}

type routeTwo struct {
}

func (two *routeTwo) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetHTML("/something", handler)
	r.Get("/posts.atom", handler)
	tracker.AddDependentUrl("/something")
}
func (two *routeTwo) WildcardUrls() ([]string, error) {
	return []string{"two", "three", "four", "five"}, nil
}

type routeThree struct {
}

func (three *routeThree) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetHTML("/about", handler)
}
func (three *routeThree) WildcardUrls() ([]string, error) {
	return nil, nil
}

func TestContent_SetRoutes(t *testing.T) {
	testCases := []struct {
		routes        []Route
		staticUrls    []string
		dependentUrls []string
	}{
		{[]Route{}, []string{}, []string{}},
		{[]Route{&routeOne{}}, []string{"/", "/about", "/posts", "/robots.txt"}, []string{"/posts", router.RootUrlPattern}},
		{[]Route{&routeTwo{}}, []string{"/something", "/posts.atom"}, []string{"/something"}},
		{[]Route{&routeThree{}}, []string{"/about"}, []string{}},
		{[]Route{&routeOne{}, &routeTwo{}}, []string{"/", "/about", "/posts", "/robots.txt", "/something", "/posts.atom"}, []string{"/posts", router.RootUrlPattern, "/something"}},
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
		tracker := app.NewTracker(func() ([]string, error) {
			return nil, nil
		})
		content.SetRoutes(r, tracker)

		got := r.StaticUrls()
		sort.Strings(got)
		sort.Strings(tc.staticUrls)
		if !cmp.Equal(got, tc.staticUrls) {
			t.Error(context.DiffString("r.StaticUrls()", got, tc.staticUrls, cmp.Diff(got, tc.staticUrls)))
		}

		got = tracker.DependentUrls()
		sort.Strings(got)
		sort.Strings(tc.dependentUrls)
		if !cmp.Equal(got, tc.dependentUrls) {
			t.Error(context.DiffString("tracker.DependentUrls()", got, tc.staticUrls, cmp.Diff(got, tc.dependentUrls)))
		}
	}
}

func TestContent_WildcardUrls(t *testing.T) {
	testCases := []struct {
		routes       []Route
		wildcardUrls []string
		error        bool
	}{
		{[]Route{}, []string{}, false},
		{[]Route{&routeOne{}}, []string{"one", "two", "three"}, false},
		{[]Route{&routeTwo{}}, []string{"two", "three", "four", "five"}, false},
		{[]Route{&routeThree{}}, []string{}, false},
		{[]Route{&routeOne{}, &routeTwo{}}, []string{"one", "two", "three", "four", "five"}, true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":  testCaseIndex,
			"routes": tc.routes,
		})

		content, log, _ := defaultContent()
		content.routes = tc.routes
		r := router.NewGenerateRouter(log)
		tracker := app.NewTracker(func() ([]string, error) {
			return nil, nil
		})
		content.SetRoutes(r, tracker)

		got, err := content.WildcardUrls()
		if err != nil {
			if !tc.error {
				t.Error(context.String(err))
			}
			continue
		}

		sort.Strings(got)
		sort.Strings(tc.wildcardUrls)
		if !cmp.Equal(got, tc.wildcardUrls) {
			t.Error(context.DiffString("content.WildcardUrls()", got, tc.wildcardUrls, cmp.Diff(got, tc.wildcardUrls)))
		}
	}
}
