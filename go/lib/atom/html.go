package atom

import (
	"strings"
	"time"
)

const EntryLimit = 100

type HtmlEntry struct {
	Id          string
	Title       string
	Updated     time.Time
	HtmlContent string
	Summary     string
	Published   time.Time
}

func (htmlEntry *HtmlEntry) ToEntry(a *Renderer) *Entry {
	return &Entry{
		ID:      strings.Join([]string{a.Settings.Host, htmlEntry.Id, htmlEntry.Published.Format("2006-01-02")}, ":"),
		Title:   htmlEntry.Title,
		Updated: htmlEntry.Updated,

		Author:  a.Author(),
		Content: &EntryContent{Content: htmlEntry.HtmlContent, Type: "html"},
		Summary: htmlEntry.Summary,
		Link:    a.AlternateLink(htmlEntry.Id),

		Published: htmlEntry.Published,
	}
}

type HtmlRenderer struct {
	Settings *Settings
}

func NewHtmlRenderer(settings *Settings) *HtmlRenderer {
	return &HtmlRenderer{settings}
}

func (renderer *HtmlRenderer) Render(feedName, selfUrl, logoUrl string, htmlEntries []*HtmlEntry) ([]byte, error) {
	atomRenderer := NewRenderer(renderer.Settings)
	feed := HtmlEntriesToFeed(atomRenderer, feedName, selfUrl, logoUrl, htmlEntries)
	return feed.Marhshall()
}

func HtmlEntriesToFeed(atomRenderer *Renderer, feedName, selfUrl, logoUrl string, htmlEntries []*HtmlEntry) *Feed {
	entries := make([]*Entry, len(htmlEntries))
	for i, htmlEntry := range htmlEntries {
		entries[i] = htmlEntry.ToEntry(atomRenderer)
	}

	lastUpdated := time.Now()
	if len(htmlEntries) >= 1 {
		lastUpdated = htmlEntries[0].Updated
	}
	feed := atomRenderer.NewFeed(feedName, lastUpdated, selfUrl, logoUrl)
	feed.Entries = entries
	return feed
}
