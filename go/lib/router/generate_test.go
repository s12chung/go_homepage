package router

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/test"
)

func defaultGenerateRouter() (*GenerateRouter, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewGenerateRouter(log), hook
}

func TestGenerateRouter_Around(t *testing.T) {
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

		router, _ := defaultGenerateRouter()
		router.GetRootHTML(func(ctx Context) error {
			testPreviousContext(ctx)

			got = append(got, "call")
			return nil
		})

		for _, handler := range tc.handlers {
			router.Around(handler)
		}

		_, err := router.Requester().Get(RootUrlPattern)
		if err != nil {
			t.Error(context.String(err))
		}
		if !cmp.Equal(got, tc.expected) {
			t.Error(context.GotExpString("state", got, tc.expected))
		}
	}
}

func testGet(t *testing.T, requestUrl string, testFunc func(router *GenerateRouter, handler ContextHandler)) {
	testRouterContext(t, requestUrl, testFunc)
	testRouterErrors(t, requestUrl, testFunc)
}

func testRouterContext(t *testing.T, requestUrl string, testFunc func(router *GenerateRouter, handler ContextHandler)) {
	called := false
	expResponse := "The Response"
	router, _ := defaultGenerateRouter()
	testFunc(router, func(ctx Context) error {
		called = true
		test.AssertLabel(t, "ctx.Log()", ctx.Log(), router.log)
		test.AssertLabel(t, "ctx.Url()", ctx.Url(), requestUrl)
		urlParts, _ := urlParts(requestUrl)
		if !cmp.Equal(ctx.UrlParts(), urlParts) {
			t.Error(test.AssertLabelString("ctx.UrlParts()", ctx.UrlParts(), urlParts))
		}

		ctx.Respond([]byte(expResponse))
		return nil
	})
	response, err := router.Requester().Get(requestUrl)
	if err != nil {
		t.Error(err)
	}
	test.AssertLabel(t, "response", string(response), expResponse)
	test.AssertLabel(t, "called", called, true)
}

func testRouterErrors(t *testing.T, requestUrl string, testFunc func(router *GenerateRouter, handler ContextHandler)) {
	called := false
	expError := "test error"
	router, _ := defaultGenerateRouter()
	testFunc(router, func(ctx Context) error {
		called = true
		return fmt.Errorf(expError)
	})
	_, err := router.Requester().Get(requestUrl)
	test.AssertLabel(t, "Handler error", err.Error(), expError)

	_, err = router.Requester().Get("/multipart/url")
	if err == nil {
		t.Error("Multipart Urls are not giving errors")
	}

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Did not panic for duplicate route setup.")
			}
		}()
		testFunc(router, func(ctx Context) error {
			return nil
		})
	}()
}

func TestGenerateRouter_GetWildcardHTML(t *testing.T) {
	testGet(t, "/anything", func(router *GenerateRouter, handler ContextHandler) {
		router.GetWildcardHTML(handler)
	})
}

func TestGenerateRouter_GetRootHTML(t *testing.T) {
	testGet(t, RootUrlPattern, func(router *GenerateRouter, handler ContextHandler) {
		router.GetRootHTML(handler)
	})
}

func TestGenerateRouter_GetHTML(t *testing.T) {
	testGet(t, "/blah", func(router *GenerateRouter, handler ContextHandler) {
		router.GetHTML("/blah", handler)
	})
}

func TestGenerateRouter_Get(t *testing.T) {
	testGet(t, "/blah.atom", func(router *GenerateRouter, handler ContextHandler) {
		router.Get("/blah.atom", "application/xml", handler)
	})
}

func TestGenerateRouter_StaticRoutes(t *testing.T) {
	handler := func(ctx Context) error {
		return nil
	}

	testCases := []struct {
		htmlRoutes  []string
		otherRoutes []string
		mimeTypes   []string
	}{
		{[]string{}, []string{}, []string{}},
		{[]string{"/some"}, []string{"/something.atom"}, []string{"application/xml"}},
		{[]string{"/some", "/ha", "/works"}, []string{"/something.atom", "/robots.txt"}, []string{"application/xml", "text/plain"}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"htmlRoutes":  tc.htmlRoutes,
			"otherRoutes": tc.otherRoutes,
		})

		router, _ := defaultGenerateRouter()
		router.GetWildcardHTML(handler)
		router.GetRootHTML(handler)

		for _, htmlRoute := range tc.htmlRoutes {
			router.GetHTML(htmlRoute, handler)
		}
		for i, route := range tc.otherRoutes {
			router.Get(route, tc.mimeTypes[i], handler)
		}

		got := router.StaticRoutes()
		exp := append(tc.htmlRoutes, tc.otherRoutes...)
		exp = append(exp, RootUrlPattern)

		sort.Strings(got)
		sort.Strings(exp)

		if !cmp.Equal(got, exp) {
			t.Error(context.GotExpString("Result", got, exp))
		}
	}
}
