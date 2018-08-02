package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Settings struct {
	GeneratedPath  string            `json:"generated_path,omitempty"`
	PostsPath      string            `json:"generated_path,omitempty"`
	DraftsPath     string            `json:"generated_path,omitempty"`
	Concurrency    int               `json:"concurrency,omitempty"`
	ServerPort     int               `json:"server_port,omitempty"`
	FileServerPort int               `json:"server_port,omitempty"`
	Template       TemplateSettings  `json:"template,omitempty"`
	Goodreads      GoodreadsSettings `json:"goodreads,omitempty"`
	Domain         DomainSettings    `json:"domain,omitempty"`
}

type TemplateSettings struct {
	WebsiteTitle   string `json:"website_title,omitempty"`
	AssetsPath     string `json:"assets_path,omitempty"`
	ManifestPath   string `json:"manifest_path,omitempty"`
	ResponsivePath string `json:"responsive_path,omitempty"`
	PostImagesPath string `json:"post_images_path,omitempty"`
	MarkdownsPath  string `json:"markdowns_path,omitempty"`
}

type GoodreadsSettings struct {
	CachePath  string `json:"cache_path,omitempty"`
	ApiKey     string `json:"api_key,omitempty"`
	UserId     int    `json:"user_id,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	MaxPerPage int    `json:"max_per_page,omitempty"`
	RateLimit  int    `json:"rate_limit,omitempty"`
}

type DomainSettings struct {
	AuthorName string `json:"author_name,omitempty"`
	AuthorUri  string `json:"author_uri,omitempty"`

	Host string `json:"host,omitempty"`
	SSL  bool   `json:"ssl,omitempty"`
}

func (domainSettings *DomainSettings) Url() string {
	ssl := ""
	if domainSettings.SSL {
		ssl = "s"
	}
	return fmt.Sprintf("http%v://%v", ssl, domainSettings.Host)
}

func (domainSettings *DomainSettings) UrlFor(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return domainSettings.Url()
	}
	return strings.Join([]string{domainSettings.Url(), path}, "/")
}

const settingsPath = "settings.json"

func ReadFromFile(log logrus.FieldLogger) *Settings {
	settings := Settings{
		"./generated",
		"./content/posts",
		"./content/drafts",
		10,
		8080,
		3000,
		TemplateSettings{
			"Your Website Title",
			"./generated/assets",
			"./generated/assets/manifest.json",
			"./generated/assets/content/responsive",
			"./generated/assets/content/images",
			"./assets/markdowns",
		},
		GoodreadsSettings{
			"cache",
			"",
			0,
			50,
			200,
			1000,
		},
		DomainSettings{
			"Your Name",
			"",
			"yourwebsite.com",
			true,
		},
	}
	_, err := os.Stat(settingsPath)
	if os.IsNotExist(err) {
		log.Infof("%v not found, using defaults...", settingsPath)
		return &settings
	}

	file, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return &settings
	}

	err = json.Unmarshal(file, &settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return &settings
	}

	settings.Domain.AuthorUri = settings.Domain.Url()
	return &settings
}
