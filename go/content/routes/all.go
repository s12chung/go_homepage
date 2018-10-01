package routes

import (
	"sort"
	"time"

	"github.com/s12chung/go_homepage/go/content/atom"
	"github.com/s12chung/go_homepage/go/content/models"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/router"

	"github.com/s12chung/gostatic-packages/goodreads"
)

type AllRoutes struct {
	h Helper
}

func NewAllRoutes(h Helper) *AllRoutes {
	return &AllRoutes{h}
}

func (routes *AllRoutes) SetRoutes(r router.Router, tracker *app.Tracker) error {
	r.GetRootHTML(routes.getPosts)
	tracker.AddDependentURL(router.RootURL)
	r.Get("/posts.atom", routes.getPostsAtom)
	tracker.AddDependentURL("/posts.atom")

	r.GetHTML("/reading", routes.getReading)
	r.GetHTML("/about", routes.getAbout)
	r.Get("/robots.txt", routes.getRobotsTxt)
	r.Get("/404.html", routes.get404)

	allPostFilenames, err := models.AllPostFilenames()
	if err != nil {
		return err
	}
	for _, filename := range allPostFilenames {
		r.GetHTML(filename, routes.getPostF(filename))
	}
	return nil
}

func (routes *AllRoutes) getAbout(ctx router.Context) error {
	return routes.h.RespondHTML(ctx, ctx.URL(), layoutData{"About", nil})
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
	return routes.h.RespondHTML(ctx, ctx.URL(), layoutData{"Reading", data})
}

func (routes *AllRoutes) getPostF(filename string) func(ctx router.Context) error {
	return func(ctx router.Context) error {
		post, err := models.NewPost(filename)
		if err != nil {
			return err
		}
		return routes.h.RespondHTML(ctx, "post", layoutData{post.Title, post})
	}
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
	return routes.h.RespondHTML(ctx, "posts", layoutData{"", data})
}

func (routes *AllRoutes) getPostsAtom(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	logoURL := routes.h.ManifestURL("images/logo.png")
	htmlEntries := atom.PostsToHTMLEntries(posts)
	return routes.h.RespondAtom(ctx, "posts", logoURL, htmlEntries)
}

func (routes *AllRoutes) getRobotsTxt(ctx router.Context) error {
	// decided not to show the directory structure via this file
	// there is a lib for robots.txt in go/lib/robots though
	ctx.Respond([]byte{})
	return nil
}

func (routes *AllRoutes) get404(ctx router.Context) error {
	return routes.h.RespondHTML(ctx, ctx.URL(), layoutData{"404", nil})
}

func sortedPosts() ([]*models.Post, error) {
	posts, err := models.Posts()
	if err != nil {
		return nil, err
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].PublishedAt.After(posts[j].PublishedAt) })
	return posts, nil
}
