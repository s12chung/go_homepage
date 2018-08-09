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
	return routeSetter.respondUrlHTML(ctx, nil)
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
	return routeSetter.respondUrlHTML(ctx, data)
}

func (routeSetter *RouteSetter) getPost(ctx router.Context) error {
	post, err := models.NewPost(ctx.UrlParts()[0])
	if err != nil {
		return err
	}
	return routeSetter.respondHTML(ctx, "post", post)
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
	return routeSetter.respondHTML(ctx, "posts", data)
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

func (routeSetter *RouteSetter) respondUrlHTML(ctx router.Context, data interface{}) error {
	return routeSetter.respondHTML(ctx, "", data)
}

func (routeSetter *RouteSetter) respondHTML(ctx router.Context, templateName string, data interface{}) error {
	bytes, err := routeSetter.renderHTML(ctx, templateName, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (routeSetter *RouteSetter) renderHTML(ctx router.Context, tmplName string, data interface{}) ([]byte, error) {
	tmplName = templateName(ctx, tmplName)
	defaultTitle := defaultTitle(ctx, tmplName)
	ctx.Log().Infof("Rendering template %v with title %v", tmplName, defaultTitle)
	return routeSetter.renderer.Render(tmplName, defaultTitle, data)
}

func templateName(ctx router.Context, templateName string) string {
	if templateName == "" {
		// assume len of <= 1: https://github.com/s12chung/go_homepage/blob/aa77eaf3ffff669b6abaab35078fb65ee3ffb17c/server/router/router.go#L52
		if router.IsRootUrlPart(ctx.UrlParts()) {
			ctx.Log().Panicf("No TemplateName given for root route")
			return ""
		}
		return ctx.UrlParts()[0]
	} else {
		return templateName
	}
}

func defaultTitle(ctx router.Context, templateName string) string {
	defaultTitle := templateName
	if router.IsRootUrlPart(ctx.UrlParts()) {
		defaultTitle = ""
	}
	return defaultTitle
}
