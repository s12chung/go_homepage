package webpack

import (
	"os"
)

type Settings struct {
	AssetsPath             string            `json:"assets_path,omitempty"`
	ReplaceResponsiveAttrs string            `json:"process_html_responsive_image,omitempty"`
	ResponsiveImageMap     map[string]string `json:"responsive_image_map,omitempty"`
}

func DefaultSettings() *Settings {
	assetsPath := os.Getenv("ASSETS_PATH")
	if assetsPath == "" {
		assetsPath = "assets"
	}
	return &Settings{
		assetsPath,
		"content",
		map[string]string{
			"assets":  "images",
			"content": "content/images",
		},
	}
}
