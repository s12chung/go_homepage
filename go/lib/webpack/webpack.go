package webpack

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var imgRegex = regexp.MustCompile(`<img (src="([^"]*)")`)
var assetsPath = ""
var assetsUrl = ""

func AssetsPath() string {
	if assetsPath != "" {
		return assetsPath
	}
	assetsPath = os.Getenv("ASSETS_PATH")
	if assetsPath != "" {
		return assetsPath
	}
	assetsPath = "assets"
	return assetsPath
}

func AssetsUrl() string {
	if assetsUrl == "" {
		assetsUrl = fmt.Sprintf("/%v/", AssetsPath())
	}
	return assetsUrl
}

type Webpack struct {
	generatedPath string
	manifest      *Manifest
	responsive    *Responsive
	log           logrus.FieldLogger
}

func NewWebpack(generatedPath string, log logrus.FieldLogger) *Webpack {
	return &Webpack{
		generatedPath,
		NewManifest(generatedPath, AssetsPath(), log),
		NewResponsive(generatedPath, AssetsPath(), "content/images", log),
		log,
	}
}

func (w *Webpack) GeneratedAssetsPath() string {
	return filepath.Join(w.generatedPath, AssetsPath())
}

func (w *Webpack) GetResponsiveImage(originalSrc string) *ResponsiveImage {
	if !HasResponsive(originalSrc) {
		manifestKey := filepath.Join(w.responsive.imagePath, filepath.Base(originalSrc))
		return &ResponsiveImage{Src: w.ManifestUrl(manifestKey)}
	}
	return w.responsive.GetResponsiveImage(originalSrc)
}

func (w *Webpack) ManifestUrl(key string) string {
	return w.manifest.ManifestUrl(key)
}

func (w *Webpack) ProcessHTML(html string) string {
	return imgRegex.ReplaceAllStringFunc(html, func(imgTag string) string {
		matches := imgRegex.FindStringSubmatch(imgTag)
		responsiveImage := w.GetResponsiveImage(matches[2])

		attributes := []string{fmt.Sprintf(`src="%v"`, responsiveImage.Src)}
		if responsiveImage.SrcSet != "" {
			attributes = append(attributes, fmt.Sprintf(`srcset="%v"`, responsiveImage.SrcSet))
		}
		return strings.Replace(imgTag, matches[1], strings.Join(attributes, " "), 1)
	})
}

func (w *Webpack) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"webpackUrl": w.ManifestUrl,
	}
}
