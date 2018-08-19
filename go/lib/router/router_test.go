package router

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/test"
	"path/filepath"
)

type RouterSetup interface {
	DefaultRouter() (Router, logrus.FieldLogger, *logTest.Hook)
	RunServer(router Router, callback func())
	Requester(router Router) Requester
}

type RouterTester struct {
	setup RouterSetup
}

func NewRouterTester(setup RouterSetup) *RouterTester {
	return &RouterTester{setup}
}

func (tester *RouterTester) TestRouter_Around(t *testing.T) {
	var got []string
	var previousContext Context

	testPreviousContext := func(ctx Context) {
		if previousContext == nil {
			previousContext = ctx
		} else {
			test.AssertLabel(t, "ctx", ctx, previousContext)
		}
	}

	h := func(before, after string) AroundHandler {
		return func(ctx Context, handler ContextHandler) error {
			testPreviousContext(ctx)

			if before != "" {
				got = append(got, before)
			}
			err := handler(ctx)
			if after != "" {
				got = append(got, after)
			}
			return err
		}
	}

	testCases := []struct {
		handlers []AroundHandler
		expected []string
	}{
		{[]AroundHandler{}, []string{"call"}},
		{[]AroundHandler{h("b1", "")}, []string{"b1", "call"}},
		{[]AroundHandler{h("b1", ""), h("b2", "")}, []string{"b1", "b2", "call"}},
		{[]AroundHandler{h("", "a1")}, []string{"call", "a1"}},
		{[]AroundHandler{h("", "a1"), h("", "a2")}, []string{"call", "a2", "a1"}},
		{[]AroundHandler{h("ar1", "ar2")}, []string{"ar1", "call", "ar2"}},
		{[]AroundHandler{h("ar1", "ar2"), h("arr1", "arr2")}, []string{"ar1", "arr1", "call", "arr2", "ar2"}},
		{[]AroundHandler{h("ar1", "ar2"), h("", "a1"), h("b1", ""), h("arr1", "arr2")}, []string{"ar1", "b1", "arr1", "call", "arr2", "a1", "ar2"}},
	}

	for testCaseIndex, tc := range testCases {
		got = nil
		previousContext = nil
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"handlersLen": len(tc.handlers),
		})

		router, _, _ := tester.setup.DefaultRouter()
		router.GetRootHTML(func(ctx Context) error {
			testPreviousContext(ctx)

			got = append(got, "call")
			return nil
		})

		for _, handler := range tc.handlers {
			router.Around(handler)
		}

		tester.setup.RunServer(router, func() {
			_, err := tester.setup.Requester(router).Get(RootUrlPattern)
			if err != nil {
				t.Error(context.String(err))
			}
			if !cmp.Equal(got, tc.expected) {
				t.Error(context.GotExpString("state", got, tc.expected))
			}
		})
	}
}

var AllGetTypesWithResponse = []struct {
	pattern  string
	mimeType string
	response string
}{
	{WildcardUrlPattern, "text/html; charset=utf-8", `<div>Being wild</div>`},
	{RootUrlPattern, "text/html; charset=utf-8", `<p>the root of it all</p>`},
	{"/page", "text/html; charset=utf-8", `<html>some page</html>`},
	{"/another_page", "text/html; charset=utf-8", `<html>another_page</html>`},
	{"/something.atom", "application/xml", `<?xml version="1.0" encoding="UTF-8"?>`},
	{"/robots.txt", "text/plain", "User-agent: *\nDisallow: /"},
}

func SetupAllGetTypesWithResponse(router Router) {
	for _, allGetTypeWithResponse := range AllGetTypesWithResponse {
		response := allGetTypeWithResponse.response
		handler := func(ctx Context) error {
			return ctx.Respond([]byte(response))
		}

		switch allGetTypeWithResponse.pattern {
		case WildcardUrlPattern:
			router.GetWildcardHTML(handler)
		case RootUrlPattern:
			router.GetRootHTML(handler)
		default:
			if filepath.Ext(allGetTypeWithResponse.pattern) == "" {
				router.GetHTML(allGetTypeWithResponse.pattern, handler)
			} else {
				router.Get(allGetTypeWithResponse.pattern, allGetTypeWithResponse.mimeType, handler)
			}
		}

	}
}

type AllGetType struct {
	htmlRoutes  []string
	otherRoutes []string
	mimeTypes   []string
}

var AllGetTypesVaried = []AllGetType{
	{[]string{}, []string{}, []string{}},
	{[]string{"/some"}, []string{"/something.atom"}, []string{"application/xml"}},
	{[]string{"/some", "/ha", "/works"}, []string{"/something.atom", "/robots.txt"}, []string{"application/xml", "text/plain"}},
}

func SetupAllGetTypeVaried(router Router, allGetType AllGetType) {
	handler := func(ctx Context) error {
		return nil
	}

	router.GetWildcardHTML(handler)
	router.GetRootHTML(handler)

	for _, htmlRoute := range allGetType.htmlRoutes {
		router.GetHTML(htmlRoute, handler)
	}
	for i, route := range allGetType.otherRoutes {
		router.Get(route, allGetType.mimeTypes[i], handler)
	}
}

func (tester *RouterTester) TestRequester_Get(t *testing.T) {
	router, _, _ := tester.setup.DefaultRouter()
	SetupAllGetTypesWithResponse(router)

	tester.setup.RunServer(router, func() {
		requeseter := tester.setup.Requester(router)
		for getIndex, allGetTypeWithResponse := range AllGetTypesWithResponse {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":    getIndex,
				"pattern":  allGetTypeWithResponse.pattern,
				"mimeType": allGetTypeWithResponse.mimeType,
				"response": allGetTypeWithResponse.response,
			})

			url := allGetTypeWithResponse.pattern
			if url == WildcardUrlPattern {
				url = "/does_not_exist"
			}

			response, err := requeseter.Get(url)
			if err != nil {
				t.Errorf(context.String(err))
			}

			got := string(response.Body)
			exp := allGetTypeWithResponse.response
			if got != exp {
				t.Error(context.GotExpString("Response.Body", got, exp))
			}

			got = response.MimeType
			exp = allGetTypeWithResponse.mimeType
			if got != exp {
				t.Error(context.GotExpString("Response.MimeType", got, exp))
			}
		}
	})
}

func (tester *RouterTester) NewGetTester(requestUrl string, testFunc func(router Router, handler ContextHandler)) *GetTester {
	return &GetTester{
		tester.setup,
		requestUrl,
		testFunc,
	}
}

type GetTester struct {
	setup      RouterSetup
	requestUrl string
	testFunc   func(router Router, handler ContextHandler)
}

func (getTester *GetTester) TestGet(t *testing.T) {
	getTester.testRouterContext(t)
	getTester.testRouterErrors(t)
}

func (getTester *GetTester) testRouterContext(t *testing.T) {
	called := false
	expResponse := "The Response"
	router, log, _ := getTester.setup.DefaultRouter()
	getTester.testFunc(router, func(ctx Context) error {
		called = true
		test.AssertLabel(t, "ctx.Log()", ctx.Log(), log)
		test.AssertLabel(t, "ctx.Url()", ctx.Url(), getTester.requestUrl)
		urlParts, _ := urlParts(getTester.requestUrl)
		if !cmp.Equal(ctx.UrlParts(), urlParts) {
			t.Error(test.AssertLabelString("ctx.UrlParts()", ctx.UrlParts(), urlParts))
		}

		return ctx.Respond([]byte(expResponse))
	})
	getTester.setup.RunServer(router, func() {
		response, err := getTester.setup.Requester(router).Get(getTester.requestUrl)
		if err != nil {
			t.Error(err)
		}
		test.AssertLabel(t, "response", string(response.Body), expResponse)
		test.AssertLabel(t, "called", called, true)
	})
}

func (getTester *GetTester) testRouterErrors(t *testing.T) {
	called := false
	expError := "test error"
	router, _, _ := getTester.setup.DefaultRouter()
	getTester.testFunc(router, func(ctx Context) error {
		called = true
		return fmt.Errorf(expError)
	})

	getTester.setup.RunServer(router, func() {
		_, err := getTester.setup.Requester(router).Get(getTester.requestUrl)
		test.AssertLabel(t, "Handler error", err.Error(), expError)

		_, err = getTester.setup.Requester(router).Get("/multipart/url")
		if err == nil {
			t.Error("Multipart Urls are not giving errors")
		}
	})
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Did not panic for duplicate route setup.")
			}
		}()
		getTester.testFunc(router, func(ctx Context) error {
			return nil
		})
	}()
}

func (tester *RouterTester) TestRouter_GetWildcardHTML(t *testing.T) {
	tester.NewGetTester("/anything", func(router Router, handler ContextHandler) {
		router.GetWildcardHTML(handler)
	}).TestGet(t)
}

func (tester *RouterTester) TestRouter_GetRootHTML(t *testing.T) {
	tester.NewGetTester(RootUrlPattern, func(router Router, handler ContextHandler) {
		router.GetRootHTML(handler)
	}).TestGet(t)
}

func (tester *RouterTester) TestRouter_GetHTML(t *testing.T) {
	tester.NewGetTester("/blah", func(router Router, handler ContextHandler) {
		router.GetHTML("/blah", handler)
	}).TestGet(t)
}

func (tester *RouterTester) TestRouter_Get(t *testing.T) {
	tester.NewGetTester("/blah.atom", func(router Router, handler ContextHandler) {
		router.Get("/blah.atom", "application/xml", handler)
	}).TestGet(t)
}

func (tester *RouterTester) TestRouter_StaticUrls(t *testing.T) {
	for testCaseIndex, allGetType := range AllGetTypesVaried {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"htmlRoutes":  allGetType.htmlRoutes,
			"otherRoutes": allGetType.otherRoutes,
		})

		router, _, _ := tester.setup.DefaultRouter()
		SetupAllGetTypeVaried(router, allGetType)

		got := router.StaticUrls()
		exp := append(allGetType.htmlRoutes, allGetType.otherRoutes...)
		exp = append(exp, RootUrlPattern)

		sort.Strings(got)
		sort.Strings(exp)

		if !cmp.Equal(got, exp) {
			t.Error(context.GotExpString("Result", got, exp))
		}
	}
}
