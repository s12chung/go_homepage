package atom

import (
	"github.com/s12chung/go_homepage/go/content/models"

	"github.com/s12chung/gostatic-packages/atom"
)

func PostsToHtmlEntries(posts []*models.Post) []*atom.HTMLEntry {
	if atom.EntryLimit < len(posts) {
		posts = posts[0:atom.EntryLimit]
	}
	htmlEntries := make([]*atom.HTMLEntry, len(posts))
	for i, post := range posts {
		htmlEntries[i] = PostToHtmlEntry(post)
	}
	return htmlEntries
}

func PostToHtmlEntry(post *models.Post) *atom.HTMLEntry {
	return &atom.HTMLEntry{
		post.Id(),
		post.Title,
		post.PublishedAt,
		post.MarkdownHTML,
		post.Description,
		post.PublishedAt,
	}
}
