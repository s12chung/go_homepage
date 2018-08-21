package webpack

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

type ResponsiveImage struct {
	SrcSet string `json:"srcSet"`
	Src    string `json:"src"`
}

var spacesRegex = regexp.MustCompile(`\s+`)

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
