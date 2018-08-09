package atom

import (
	"strings"
	"time"

	"github.com/s12chung/go_homepage/go/lib/view/atom"
)

const atomPostLimit = 100

type HtmlEntry struct {
	Id          string
	Title       string
	Updated     time.Time
	HtmlContent string
	Summary     string
	Published   time.Time
}

func (htmlEntry *HtmlEntry) ToEntry(a *atom.AtomRenderer) *atom.Entry {
	return &atom.Entry{
		ID:      strings.Join([]string{a.Settings.Host, htmlEntry.Id, htmlEntry.Published.Format("2006-01-02")}, ":"),
		Title:   htmlEntry.Title,
		Updated: htmlEntry.Updated,

		Author:  a.Author(),
		Content: &atom.EntryContent{Content: htmlEntry.HtmlContent, Type: "html"},
		Summary: htmlEntry.Summary,
		Link:    a.AlternateLink(htmlEntry.Id),

		Published: htmlEntry.Published,
	}
}

func Render(settings *atom.Settings, entryName, url, logoPath string, htmlEntries []*HtmlEntry) ([]byte, error) {
	atomRenderer := atom.NewAtomRenderer(settings)
	feed := HtmlEntriesToFeed(atomRenderer, entryName, url, logoPath, htmlEntries)
	return feed.Marhshall()
}

func HtmlEntriesToFeed(atomRenderer *atom.AtomRenderer, entryName, url, logoPath string, htmlEntries []*HtmlEntry) *atom.Feed {
	entries := make([]*atom.Entry, len(htmlEntries))
	for i, htmlEntry := range htmlEntries {
		entries[i] = htmlEntry.ToEntry(atomRenderer)
	}

	lastUpdated := time.Now()
	if len(htmlEntries) >= 1 {
		lastUpdated = htmlEntries[0].Updated
	}
	feed := atomRenderer.NewFeed(entryName, lastUpdated, url, logoPath)
	feed.Entries = entries
	return feed
}
