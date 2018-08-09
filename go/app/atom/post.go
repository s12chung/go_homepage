package atom

import (
	"github.com/s12chung/go_homepage/go/app/models"
)

func PostsToHtmlEntries(posts []*models.Post) []*HtmlEntry {
	if atomPostLimit < len(posts) {
		posts = posts[0 : atomPostLimit-1]
	}
	htmlEntries := make([]*HtmlEntry, len(posts))
	for i, post := range posts {
		htmlEntries[i] = PostToHtmlEntry(post)
	}
	return htmlEntries
}

func PostToHtmlEntry(post *models.Post) *HtmlEntry {
	return &HtmlEntry{
		post.Id(),
		post.Title,
		post.PublishedAt,
		post.MarkdownHTML,
		post.Description,
		post.PublishedAt,
	}
}
