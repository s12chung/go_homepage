package atom

import (
	"github.com/s12chung/go_homepage/go/content/models"

	"github.com/s12chung/gostatic-packages/atom"
)

func PostsToHTMLEntries(posts []*models.Post) []*atom.HTMLEntry {
	if atom.EntryLimit < len(posts) {
		posts = posts[0:atom.EntryLimit]
	}
	htmlEntries := make([]*atom.HTMLEntry, len(posts))
	for i, post := range posts {
		htmlEntries[i] = PostToHTMLEntry(post)
	}
	return htmlEntries
}

func PostToHTMLEntry(post *models.Post) *atom.HTMLEntry {
	return &atom.HTMLEntry{
		ID:          post.ID(),
		Title:       post.Title,
		Updated:     post.PublishedAt,
		HTMLContent: post.MarkdownHTML,
		Summary:     post.Description,
		Published:   post.PublishedAt,
	}
}
