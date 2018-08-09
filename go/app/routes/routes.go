package routes

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/s12chung/go_homepage/go/app/atom"
	"github.com/s12chung/go_homepage/go/app/models"
	"github.com/s12chung/go_homepage/go/app/settings"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/router"
)

var dependentUrls = map[string]bool{
	"/":           true,
	"/posts.atom": true,
}

type RouteSetter struct {
	Router   router.Router
	renderer *html.Renderer
	settings *settings.Settings
}

func NewRouteSetter(r router.Router, renderer *html.Renderer, s *settings.Settings) *RouteSetter {
	return &RouteSetter{r, renderer, s}
}

func (routeSetter *RouteSetter) SetRoutes() {
	routeSetter.Router.GetRootHTML(routeSetter.getPosts)
	routeSetter.Router.GetHTML("/posts.atom", routeSetter.getPostsAtom)

	routeSetter.Router.GetWildcardHTML(routeSetter.getPost)
	routeSetter.Router.GetHTML("/reading", routeSetter.getReading)
	routeSetter.Router.GetHTML("/about", routeSetter.getAbout)
	routeSetter.Router.GetHTML("/robots.txt", routeSetter.getRobotsTxt)
}

func (routeSetter *RouteSetter) AllUrls() ([]string, error) {
	allUrls := routeSetter.Router.StaticRoutes()
	allPostFilenames, err := models.AllPostFilenames()
	if err != nil {
		return nil, err
	}

	hasSpace := regexp.MustCompile(`\s`).MatchString
	for i, filename := range allPostFilenames {
		if hasSpace(filename) {
			return nil, fmt.Errorf("filename '%v' has a space", filename)
		}
		allPostFilenames[i] = "/" + filename
	}

	return append(allUrls, allPostFilenames...), nil
}

func (routeSetter *RouteSetter) IndependentUrls() ([]string, error) {
	allUrls, err := routeSetter.AllUrls()
	if err != nil {
		return nil, err
	}

	independentUrls := make([]string, len(allUrls)-len(dependentUrls))
	i := 0
	for _, url := range allUrls {
		_, exists := dependentUrls[url]
		if !exists {
			independentUrls[i] = url
			i += 1
		}
	}
	return independentUrls, nil
}

func (routeSetter *RouteSetter) DependentUrls() []string {
	urls := make([]string, len(dependentUrls))
	i := 0
	for url := range dependentUrls {
		urls[i] = url
		i += 1
	}
	return urls
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
