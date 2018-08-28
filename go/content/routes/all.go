package routes

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/s12chung/go_homepage/go/content/atom"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/goodreads"
	"github.com/s12chung/gostatic/go/lib/router"
)

type AllRoutes struct {
	h Helper
}

func NewAllRoutes(h Helper) *AllRoutes {
	return &AllRoutes{h}
}

func (routes *AllRoutes) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetRootHTML(routes.getPosts)
	tracker.AddDependentUrl(router.RootUrlPattern)
	r.Get("/posts.atom", routes.getPostsAtom)
	tracker.AddDependentUrl("/posts.atom")

	r.GetWildcardHTML(routes.getPost)
	r.GetHTML("/reading", routes.getReading)
	r.GetHTML("/about", routes.getAbout)
	r.Get("/robots.txt", routes.getRobotsTxt)
}

func (routes *AllRoutes) WildcardUrls() ([]string, error) {
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
	return allPostFilenames, nil
}

func (routes *AllRoutes) getAbout(ctx router.Context) error {
	return routes.h.RespondUrlHTML(ctx, nil)
}

type readingData struct {
	Books        []*goodreads.Book
	RatingMap    map[int]int
	EarliestYear int
}

func (routes *AllRoutes) getReading(ctx router.Context) error {
	books, err := goodreads.NewClient(routes.h.GoodreadsSettings(), ctx.Log()).GetBooks()
	if err != nil {
		return err
	}
	sort.Slice(books, func(i, j int) bool { return books[i].SortedDate().After(books[j].SortedDate()) })

	earliestYear := time.Now().Year()
	if len(books) >= 1 {
		earliestYear = books[len(books)-1].SortedDate().Year()
	}

	data := readingData{
		books,
		goodreads.RatingMap(books),
		earliestYear,
	}
	return routes.h.RespondUrlHTML(ctx, data)
}

func (routes *AllRoutes) getPost(ctx router.Context) error {
	post, err := models.NewPost(ctx.UrlParts()[0])
	if err != nil {
		return err
	}
	return routes.h.RespondHTML(ctx, "post", post)
}

type postsData struct {
	Posts []*models.Post
}

func (routes *AllRoutes) getPosts(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	data := postsData{
		posts,
	}
	return routes.h.RespondHTML(ctx, "posts", data)
}

func (routes *AllRoutes) getPostsAtom(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	logoUrl := routes.h.ManifestUrl("images/logo.png")
	htmlEntries := atom.PostsToHtmlEntries(posts)
	return routes.h.RespondAtom(ctx, "posts", logoUrl, htmlEntries)
}

func (routes *AllRoutes) getRobotsTxt(ctx router.Context) error {
	// decided not to show the directory structure via this file
	// there is a lib for robots.txt in go/lib/robots though
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
