package router

import (
	"net/http/httptest"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/sirupsen/logrus"
	"net/url"
	"strconv"
)

func defaultWebRouter() (*WebRouter, logrus.FieldLogger, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewWebRouter(8080, log), log, hook
}

func webRouterTester() *RouterTester {
	return NewRouterTester(&WebRouterSetup{})
}

type WebRouterSetup struct {
	server *httptest.Server
}

func (setup *WebRouterSetup) DefaultRouter() (Router, logrus.FieldLogger, *logTest.Hook) {
	return defaultWebRouter()
}

func (setup *WebRouterSetup) RunServer(router Router, callback func()) {
	r, ok := router.(*WebRouter)
	if !ok {
		panic("Not a *WebRouter being passed")
	}
	setup.server = httptest.NewServer(r.serveMux)
	callback()
	setup.server.Close()
}

func (setup *WebRouterSetup) Requester(router Router) Requester {
	urlObject, err := url.Parse(setup.server.URL)
	if err != nil {
		panic(err)
	}
	port, err := strconv.ParseInt(urlObject.Port(), 10, 32)
	if err != nil {
		panic(err)
	}

	requester := newWebRequester(int(port))
	requester.hostname = urlObject.Hostname()
	return requester
}

func TestWebRouter_Around(t *testing.T) {
	webRouterTester().TestRouter_Around(t)
}

func TestWebRouter_GetWildcardHTML(t *testing.T) {
	webRouterTester().TestRouter_GetWildcardHTML(t)
}

func TestWebRouter_GetRootHTML(t *testing.T) {
	webRouterTester().TestRouter_GetRootHTML(t)
}

func TestWebRouter_GetHTML(t *testing.T) {
	webRouterTester().TestRouter_GetHTML(t)
}

func TestWebRouter_Get(t *testing.T) {
	webRouterTester().TestRouter_Get(t)
}

func TestWebRouter_StaticRoutes(t *testing.T) {
	webRouterTester().TestRouter_StaticRoutes(t)
}
