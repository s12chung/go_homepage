package routes

import (
	"sort"

	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/server/router"
)

var DependentUrls = map[string]bool{
	"/": true,
}

func SetRoutes(r router.Router) {
	r.GetRootHTML(getPosts)
	r.GetWildcardHTML(getPost)
	r.GetHTML("/reading", getReading)
	r.GetHTML("/about", getAbout)
}

func getAbout(ctx router.Context) error {
	return ctx.Render(nil)
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
	return ctx.Render(data)
}

func getPost(ctx router.Context) error {
	filename := ctx.UrlParts()[0]
	post, err := models.NewPost(filename)
	if err != nil {
		return err
	}
	ctx.SetTemplateName("post")
	return ctx.Render(post)
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
	ctx.SetTemplateName("posts")
	return ctx.Render(data)
}
