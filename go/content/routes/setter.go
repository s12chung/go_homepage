package routes

import (
	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/lib/router"
)

type Setter struct {
	h *app.RespondHelper
}

func NewSetter(h *app.RespondHelper) *Setter {
	return &Setter{h}
}

func (setter *Setter) SetRoutes(r router.Router, tracker *app.Tracker) {
	setter.setAllRoutes(r, tracker)
}

func (setter *Setter) AssetsUrl() string {
	return setter.h.Renderer.AssetsUrl()
}

func (setter *Setter) GeneratedAssetsPath() string {
	return setter.h.Renderer.GeneratedAssetsPath()
}

func (setter *Setter) WildcardRoutes() ([]string, error) {
	return setter.WildcardPostRoutes()
}
