package atom

import (
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/gostatic/go/lib/atom"
)

func PostsToHtmlEntries(posts []*models.Post) []*atom.HtmlEntry {
	if atom.EntryLimit < len(posts) {
		posts = posts[0:atom.EntryLimit]
	}
	htmlEntries := make([]*atom.HtmlEntry, len(posts))
	for i, post := range posts {
		htmlEntries[i] = PostToHtmlEntry(post)
	}
	return htmlEntries
}

func PostToHtmlEntry(post *models.Post) *atom.HtmlEntry {
	return &atom.HtmlEntry{
		post.Id(),
		post.Title,
		post.PublishedAt,
		post.MarkdownHTML,
		post.Description,
		post.PublishedAt,
	}
}
