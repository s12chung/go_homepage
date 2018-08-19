package app

import "github.com/s12chung/go_homepage/go/lib/router"

//go:generate mockgen -destination=../test/mocks/setter_mock.go -package=mocks github.com/s12chung/go_homepage/go/app Setter
type Setter interface {
	SetRoutes(r router.Router, tracker *Tracker)
	WildcardUrls() ([]string, error)
	AssetsUrl() string
	GeneratedAssetsPath() string
}
