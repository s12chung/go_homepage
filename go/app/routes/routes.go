package routes

import (
	"sort"
	"time"

	"github.com/s12chung/go_homepage/go/app/atom"
	"github.com/s12chung/go_homepage/go/app/models"
	"github.com/s12chung/go_homepage/go/app/settings"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/server/router"
	"github.com/s12chung/go_homepage/go/lib/view"
)

var DependentUrls = map[string]bool{
	"/": true,
}

type RouteSetter struct {
	router   router.Router
	renderer *view.Renderer
	settings *settings.Settings
}

func NewRouteSetter(r router.Router, renderer *view.Renderer, s *settings.Settings) *RouteSetter {
	return &RouteSetter{r, renderer, s}
}

func (routeSetter *RouteSetter) SetRoutes() {
	routeSetter.router.GetRootHTML(routeSetter.getPosts)
	routeSetter.router.GetWildcardHTML(routeSetter.getPost)
	routeSetter.router.GetHTML("/reading", routeSetter.getReading)
	routeSetter.router.GetHTML("/about", routeSetter.getAbout)

	routeSetter.router.GetHTML("/posts.atom", routeSetter.getPostsAtom)
	routeSetter.router.GetHTML("/robots.txt", routeSetter.getRobotsTxt)
}

func (routeSetter *RouteSetter) getAbout(ctx router.Context) error {
	return ctx.Render(nil)
}

func (routeSetter *RouteSetter) getReading(ctx router.Context) error {
	books, err := goodreads.NewClient(&routeSetter.settings.Goodreads, ctx.Log()).GetBooks()
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

func (routeSetter *RouteSetter) getPost(ctx router.Context) error {
	post, err := models.NewPost(ctx.UrlParts()[0])
	if err != nil {
		return err
	}
	ctx.SetTemplateName("post")
	return ctx.Render(post)
}

func (routeSetter *RouteSetter) getPosts(ctx router.Context) error {
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

func (routeSetter *RouteSetter) getPostsAtom(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	logoPath := routeSetter.renderer.Webpack().ManifestPath("images/logo.png")
	htmlEntries := atom.PostsToHtmlEntries(posts)

	bytes, err := atom.Render(&routeSetter.settings.Atom, "posts", ctx.Url(), logoPath, htmlEntries)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (routeSetter *RouteSetter) getRobotsTxt(ctx router.Context) error {
	return ctx.Respond([]byte{})
}

func sortedPosts() ([]*models.Post, error) {
	posts, err := models.Posts()
	if err != nil {
		return nil, err
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].PublishedAt.After(posts[j].PublishedAt) })
	return posts, nil
}
