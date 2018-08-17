package router

import (
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/sirupsen/logrus"
)

func defaultGenerateRouter() (*GenerateRouter, logrus.FieldLogger, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewGenerateRouter(log), log, hook
}

func generateRouterTester() *RouterTester {
	return &RouterTester{func() (Router, logrus.FieldLogger, *logTest.Hook) {
		return defaultGenerateRouter()
	}}
}

func TestGenerateRouter_Around(t *testing.T) {
	generateRouterTester().TestRouter_Around(t)
}

func TestGenerateRouter_GetWildcardHTML(t *testing.T) {
	generateRouterTester().TestRouter_GetWildcardHTML(t)
}

func TestGenerateRouter_GetRootHTML(t *testing.T) {
	generateRouterTester().TestRouter_GetRootHTML(t)
}

func TestGenerateRouter_GetHTML(t *testing.T) {
	generateRouterTester().TestRouter_GetHTML(t)
}

func TestGenerateRouter_Get(t *testing.T) {
	generateRouterTester().TestRouter_Get(t)
}

func TestGenerateRouter_StaticRoutes(t *testing.T) {
	generateRouterTester().TestRouter_StaticRoutes(t)
}
