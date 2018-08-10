package router

type Setter interface {
	SetRoutes(r Router, tracker *Tracker)
	WildcardRoutes() ([]string, error)
	AssetsUrl() string
	GeneratedAssetsPath() string
}
