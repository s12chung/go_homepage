package webpack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/settings"
)

// in settings.Settings, GeneratedPath must have:
const AssetsPath = "assets"

const manifestpath = "manifest.json"

func fullManifestPath() string {
	return fmt.Sprintf("%v/%v", AssetsPath, manifestpath)
}

const postImagesPath = "content/images"

func fullPostImagesPath() string {
	return fmt.Sprintf("%v/%v", AssetsPath, postImagesPath)
}

const responsivePath = "content/responsive"

func fullResponsivePath() string {
	return fmt.Sprintf("%v/%v", AssetsPath, responsivePath)
}

var responsiveExtensions = map[string]bool{
	".png": true,
	".jpg": true,
}

var spacesRegex = regexp.MustCompile(`\s+`)

type Webpack struct {
	generatedPath string
	settings      *settings.TemplateSettings
	manifestMap   map[string]string
	log           logrus.FieldLogger
}

func NewWebpack(generatedPath string, templateSettings *settings.TemplateSettings, log logrus.FieldLogger) *Webpack {
	return &Webpack{
		generatedPath,
		templateSettings,
		map[string]string{},
		log,
	}
}

type ResponsiveImage struct {
	SrcSet string `json:"srcSet"`
	Src    string `json:"src"`
}

func (r *ResponsiveImage) changeResponsiveImageUrl(imagesUrl string) error {
	r.Src = r.changeSrc(imagesUrl, r.Src)
	if r.SrcSet == "" {
		return nil
	}

	srcWidths := strings.Split(r.SrcSet, ",")
	newSrcSet := make([]string, len(srcWidths))
	for i, srcWidth := range srcWidths {
		srcWidthSplit := spacesRegex.Split(strings.Trim(srcWidth, " "), -1)
		if len(srcWidthSplit) != 2 {
			return fmt.Errorf("srcSet is not formatted correctly with '%v' for img src='%v'", srcWidth, r.Src)
			newSrcSet[i] = srcWidth
		}
		newSrcSet[i] = fmt.Sprintf("%v %v", r.changeSrc(imagesUrl, srcWidthSplit[0]), srcWidthSplit[1])
	}

	r.SrcSet = strings.Join(newSrcSet, ", ")
	return nil
}

func (r *ResponsiveImage) changeSrc(imagesUrl, src string) string {
	return fmt.Sprintf("%v/%v", imagesUrl, path.Base(src))
}

func (w *Webpack) GetResponsiveImage(originalSrc string) *ResponsiveImage {
	responsiveImage, err := w.getResponsiveImage(originalSrc)
	if err != nil {
		w.log.Errorf("GetResponsiveImage error: %v", err)
		return &ResponsiveImage{Src: originalSrc}
	}
	return responsiveImage
}

func (w *Webpack) getResponsiveImage(originalSrc string) (*ResponsiveImage, error) {
	_, hasResponsive := responsiveExtensions[filepath.Ext(originalSrc)]
	if !hasResponsive {
		manifestKey := filepath.Join(postImagesPath, filepath.Base(originalSrc))
		return &ResponsiveImage{Src: w.ManifestPath(manifestKey)}, nil
	}

	u, err := url.Parse(originalSrc)
	if err != nil {
		return nil, err
	}
	if u.Hostname() != "" {
		return &ResponsiveImage{Src: originalSrc}, nil
	}

	responsiveImage, err := w.readResponsiveImageJSON(originalSrc)
	if err != nil {
		return nil, err
	}
	err = responsiveImage.changeResponsiveImageUrl(fullPostImagesPath())
	if err != nil {
		return nil, err
	}

	return responsiveImage, nil
}

func (w *Webpack) readResponsiveImageJSON(originalSrc string) (*ResponsiveImage, error) {
	responsiveImageFilename := fmt.Sprintf("%v.json", filepath.Base(originalSrc))
	bytes, err := ioutil.ReadFile(path.Join(w.generatedPath, fullResponsivePath(), responsiveImageFilename))
	if err != nil {
		return nil, err
	}

	responsiveImage := &ResponsiveImage{}
	err = json.Unmarshal(bytes, responsiveImage)
	if err != nil {
		return nil, err
	}
	return responsiveImage, nil
}

func (w *Webpack) manifestValue(key string) string {
	if len(w.manifestMap) == 0 {
		err := w.readManifest()
		if err != nil {
			w.log.Errorf("readManifest error: %v", err)
			return ""
		}
	}
	value := w.manifestMap[key]
	if value == "" {
		w.log.Errorf("webpack manifestValue not found for key: %v", key)
	}
	return value
}

func (w *Webpack) readManifest() error {
	bytes, err := ioutil.ReadFile(filepath.Join(w.generatedPath, fullManifestPath()))
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &w.manifestMap)
}

func (w *Webpack) ManifestPath(key string) string {
	return AssetsPath + "/" + w.manifestValue(key)
}
