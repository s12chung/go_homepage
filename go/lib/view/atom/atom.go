package atom

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type Feed struct {
	XMLName xml.Name `xml:"feed"`

	XMLLang string `xml:"xml:lang,attr"`
	XMLNS   string `xml:"xmlns,attr"`

	ID      string    `xml:"id"`
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`

	Icon   string  `xml:"icon"`
	Author *Author `xml:"author"`

	Links   []*Link  `xml:"link"`
	Entries []*Entry `xml:"entry"`
}

func (feed *Feed) Marhshall() ([]byte, error) {
	bytes, err := xml.MarshalIndent(&feed, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), bytes...), nil
}

type Entry struct {
	XMLName xml.Name `xml:"entry"`

	ID      string    `xml:"id"`
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`

	Author  *Author       `xml:"author"`
	Content *EntryContent `xml:"content"`
	Link    *Link         `xml:"link"`
	Summary string        `xml:"summary"`

	Published time.Time `xml:"published"`
}

type Author struct {
	XMLName xml.Name `xml:"author"`

	Name string `xml:"name,omitempty"`
	Uri  string `xml:"uri,omitempty"`
}

type EntryContent struct {
	XMLName xml.Name `xml:"content"`

	Content string `xml:",cdata"`
	Type    string `xml:"type,attr,omitempty"`
}

type Link struct {
	XMLName xml.Name `xml:"link"`

	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Href string `xml:"href,attr"`
}

type AtomRenderer struct {
	Settings *Settings
}

func NewAtomRenderer(settings *Settings) *AtomRenderer {
	return &AtomRenderer{settings}
}

func (a *AtomRenderer) Author() *Author {
	return &Author{Name: a.Settings.AuthorName, Uri: a.Settings.AuthorUri}
}

func (a *AtomRenderer) AlternateLink(path string) *Link {
	return &Link{Rel: "alternate", Type: "text/html", Href: a.Settings.UrlFor(path)}
}

func (a *AtomRenderer) NewFeed(entryName string, lastUpdated time.Time, selfPath, iconPath string) *Feed {
	return &Feed{
		XMLLang: "en-US",
		XMLNS:   "http://www.w3.org/2005/Atom",

		Title:   fmt.Sprintf("%v - %v", strings.Title(entryName), a.Settings.Host),
		Icon:    a.Settings.UrlFor(iconPath),
		ID:      strings.Join([]string{a.Settings.Host, "2018", entryName}, ":"),
		Updated: lastUpdated,

		Author: a.Author(),

		Links: []*Link{
			{Rel: "self", Type: "application/atom+xml", Href: a.Settings.UrlFor(selfPath)},
			a.AlternateLink(""),
		},
	}
}
