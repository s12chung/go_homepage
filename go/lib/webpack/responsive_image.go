package webpack

import (
	"fmt"
	"github.com/s12chung/go_homepage/go/lib/utils"
	"github.com/sirupsen/logrus"
	"path"
	"regexp"
	"strings"
)

type ResponsiveImage struct {
	Src    string `json:"src"`
	SrcSet string `json:"srcSet"`
}

var spacesRegex = regexp.MustCompile(`\s+`)

func (r *ResponsiveImage) ChangeSrcPrefix(prefix string, log logrus.FieldLogger) {
	r.Src = changeSrcPrefix(prefix, r.Src)
	if r.SrcSet == "" {
		return
	}

	var newSrcSet []string
	for _, srcWidth := range strings.Split(r.SrcSet, ",") {
		srcWidthSplit := spacesRegex.Split(strings.Trim(srcWidth, " "), -1)
		if len(srcWidthSplit) != 2 {
			log.Warn("skipping, srcSet is not formatted correctly with '%v' for img src='%v'", srcWidth, r.Src)
			continue
		}
		newSrcSet = append(newSrcSet, fmt.Sprintf("%v %v", changeSrcPrefix(prefix, srcWidthSplit[0]), srcWidthSplit[1]))
	}

	r.SrcSet = strings.Join(newSrcSet, ", ")
}

func (r *ResponsiveImage) HtmlAttrs() string {
	var htmlAttrs []string
	if r.Src != "" {
		htmlAttrs = append(htmlAttrs, fmt.Sprintf(`src="%v"`, r.Src))
	}
	if r.SrcSet != "" {
		htmlAttrs = append(htmlAttrs, fmt.Sprintf(`srcset="%v"`, r.SrcSet))
	}
	return strings.Join(htmlAttrs, " ")
}

func changeSrcPrefix(prefix, src string) string {
	if src == "" {
		return ""
	}

	prefix = utils.CleanFilePath(prefix)
	base := path.Base(src)
	if prefix == "" {
		return base
	}
	return fmt.Sprintf("%v/%v", prefix, base)
}
