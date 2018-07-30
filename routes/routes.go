package routes

import (
	"sort"

	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/server/router"
)

var DependentUrls = map[string]bool{
	"/posts": true,
}

func SetRoutes(r router.Router) {
	r.GetRootHTML(getIndex)
	r.GetHTML("/reading", getReading)
	r.GetHTML("/posts", getPosts)
	r.GetWildcardHTML(getPost)
}

func getIndex(ctx router.Context) error {
	return ctx.Render("index", nil)
}

func getReading(ctx router.Context) error {
	books, err := goodreads.NewClient(&ctx.Settings().Goodreads, ctx.Log()).GetBooks()
	if err != nil {
		return err
	}
	sort.Slice(books, func(i, j int) bool { return books[i].SortedDate().After(books[j].SortedDate()) })

	data := struct {
		Books        []*goodreads.Book
		RatingMap    map[int]int
		EarliestYear int
	}{
		books,
		goodreads.RatingMap(books),
		books[len(books)-1].SortedDate().Year(),
	}
	return ctx.Render("reading", data)
}

func getPost(ctx router.Context) error {
	filename := ctx.UrlParts()[0]
	post, err := models.NewPost(filename)
	if err != nil {
		return err
	}
	return ctx.Render("post", post)
}

func getPosts(ctx router.Context) error {
	posts, err := models.Posts(func(post *models.Post) bool {
		return !post.IsDraft
	})
	if err != nil {
		return err
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].PublishedAt.After(posts[j].PublishedAt) })

	data := struct {
		Posts []*models.Post
	}{
		posts,
	}
	return ctx.Render("posts", data)
}
