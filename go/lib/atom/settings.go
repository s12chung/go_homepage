package atom

import (
	"fmt"
	"strings"
)

func DefaultSettings() *Settings {
	return &Settings{
		"Your Name",
		"",
		"yourwebsite.com",
		true,
	}
}

type Settings struct {
	AuthorName string `json:"author_name,omitempty"`
	AuthorURI  string `json:"author_uri,omitempty"`

	Host string `json:"host,omitempty"`
	SSL  bool   `json:"ssl,omitempty"`
}

func (domainSettings *Settings) AuthorURIDefaulted() string {
	if domainSettings.AuthorURI == "" {
		return domainSettings.Url()
	}
	return domainSettings.AuthorURI
}

func (domainSettings *Settings) Url() string {
	ssl := ""
	if domainSettings.SSL {
		ssl = "s"
	}
	return fmt.Sprintf("http%v://%v", ssl, domainSettings.Host)
}

func (domainSettings *Settings) FullUrlFor(url string) string {
	url = strings.Trim(url, "/")
	if url == "" {
		return domainSettings.Url()
	}
	return strings.Join([]string{domainSettings.Url(), url}, "/")
}
