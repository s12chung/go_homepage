package webpack

import (
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var imgRegex = regexp.MustCompile(`<img[^>]+(src="([^"]*)")`)

type Webpack struct {
	generatedPath string
	settings      *Settings
	manifest      *Manifest
	responsiveMap map[string]*Responsive
	log           logrus.FieldLogger
}

func NewWebpack(generatedPath string, settings *Settings, log logrus.FieldLogger) *Webpack {
	responsiveMap := map[string]*Responsive{}
	for k, imagePath := range settings.ResponsiveImageMap {
		responsiveMap[k] = NewResponsive(generatedPath, settings.AssetsPath, imagePath, log)
	}

	return &Webpack{
		generatedPath,
		settings,
		NewManifest(generatedPath, settings.AssetsPath, log),
		responsiveMap,
		log,
	}
}

func (w *Webpack) AssetsPath() string {
	return w.settings.AssetsPath
}

func (w *Webpack) AssetsUrl() string {
	return fmt.Sprintf("/%v/", w.AssetsPath())
}

func (w *Webpack) GeneratedAssetsPath() string {
	return filepath.Join(w.generatedPath, w.AssetsPath())
}

func (w *Webpack) ManifestUrl(key string) string {
	return w.manifest.ManifestUrl(key)
}

func (w *Webpack) GetResponsiveImage(key, originalSrc string) *ResponsiveImage {
	responsive, has := w.responsiveMap[key]
	if !has {
		w.log.Errorf("Invalid key given to GetResponsiveImage: %v", key)
		return &ResponsiveImage{Src: originalSrc}
	}
	if !HasResponsive(originalSrc) {
		manifestKey := filepath.Join(responsive.imagePath, filepath.Base(originalSrc))
		return &ResponsiveImage{Src: w.ManifestUrl(manifestKey)}
	}
	return responsive.GetResponsiveImage(originalSrc)
}

func (w *Webpack) ReplaceResponsiveAttrs(html string) string {
	responsiveKey := w.settings.ReplaceResponsiveAttrs
	if w.settings.ReplaceResponsiveAttrs == "" {
		w.log.Errorf("no settings.ReplaceResponsiveAttrs found")
		return html
	}
	return imgRegex.ReplaceAllStringFunc(html, func(imgTag string) string {
		matches := imgRegex.FindStringSubmatch(imgTag)
		responsiveImage := w.GetResponsiveImage(responsiveKey, matches[2])
		return strings.Replace(imgTag, matches[1], responsiveImage.HtmlAttrs(), 1)
	})
}

func (w *Webpack) ResponsiveHtmlAttrs(key, originalSrc string) template.HTMLAttr {
	responsiveImage := w.GetResponsiveImage(key, originalSrc)
	return template.HTMLAttr(responsiveImage.HtmlAttrs())
}

func (w *Webpack) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"webpackUrl":             w.ManifestUrl,
		"responsiveAttrs":        w.ResponsiveHtmlAttrs,
		"replaceResponsiveAttrs": w.ReplaceResponsiveAttrs,
	}
}
