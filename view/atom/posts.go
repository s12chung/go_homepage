package atom

import (
	"time"

	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/server/router"
	"strings"
)

func (a *AtomRenderer) PostsToFeed(ctx router.Context, sortedPosts []*models.Post) *Feed {
	entries := make([]*Entry, len(sortedPosts))
	for i, post := range sortedPosts {
		entries[i] = a.PostToEntry(post)
	}

	lastUpdated := time.Now()
	if len(sortedPosts) >= 1 {
		lastUpdated = sortedPosts[0].PublishedAt
	}
	logoPath := ctx.Renderer().Webpack().ManifestPath("images/logo.png")

	feed := a.NewFeed("posts", lastUpdated, ctx.Url(), logoPath)
	feed.Entries = entries
	return feed
}

func (a *AtomRenderer) PostToEntry(post *models.Post) *Entry {
	return &Entry{
		ID:      strings.Join([]string{a.settings.Host, post.Id(), post.PublishedAt.Format("2006-01-02")}, ":"),
		Title:   post.Title,
		Updated: post.PublishedAt,

		Author:  a.author(),
		Content: &EntryContent{Content: post.MarkdownHTML, Type: "html"},
		Summary: post.Description,
		Link:    a.alternateLink(post.Id()),

		Published: post.PublishedAt,
	}
}
