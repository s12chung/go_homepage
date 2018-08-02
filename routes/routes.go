package routes

import (
	"sort"
	"time"

	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/server/router"
	"github.com/s12chung/go_homepage/view/atom"
)

var DependentUrls = map[string]bool{
	"/": true,
}

func SetRoutes(r router.Router) {
	r.GetRootHTML(getPosts)
	r.GetWildcardHTML(getPost)
	r.GetHTML("/reading", getReading)
	r.GetHTML("/about", getAbout)

	r.GetHTML("/posts.atom", getPostsAtom)
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

	earliestYear := time.Now().Year()
	if len(books) >= 1 {
		earliestYear = books[len(books)-1].SortedDate().Year()
	}

	data := struct {
		Books        []*goodreads.Book
		RatingMap    map[int]int
		EarliestYear int
	}{
		books,
		goodreads.RatingMap(books),
		earliestYear,
	}
	return ctx.Render(data)
}

func getPost(ctx router.Context) error {
	post, err := models.NewPost(ctx.UrlParts()[0])
	if err != nil {
		return err
	}
	ctx.SetTemplateName("post")
	return ctx.Render(post)
}

func getPosts(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	data := struct {
		Posts []*models.Post
	}{
		posts,
	}
	ctx.SetTemplateName("posts")
	return ctx.Render(data)
}

func getPostsAtom(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	limit := 100
	if limit > len(posts) {
		limit = len(posts)
	}
	posts = posts[0 : limit-1]

	atomRenderer := atom.NewAtomRenderer(&ctx.Settings().Domain)
	bytes, err := atomRenderer.PostsToFeed(ctx, posts).Marhshall()
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func sortedPosts() ([]*models.Post, error) {
	posts, err := models.Posts()
	if err != nil {
		return nil, err
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].PublishedAt.After(posts[j].PublishedAt) })
	return posts, nil
}
