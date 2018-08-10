package routes

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/s12chung/go_homepage/go/content/atom"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/router"
)

func (setter *Setter) setAllRoutes(r router.Router, tracker *router.Tracker) {
	r.GetRootHTML(setter.getPosts)
	tracker.AddDependentUrl(router.RootUrlPattern)
	r.GetHTML("/posts.atom", setter.getPostsAtom)
	tracker.AddDependentUrl("/posts.atom")

	r.GetWildcardHTML(setter.getPost)
	r.GetHTML("/reading", setter.getReading)
	r.GetHTML("/about", setter.getAbout)
	r.GetHTML("/robots.txt", setter.getRobotsTxt)
}

func (setter *Setter) WildcardPostRoutes() ([]string, error) {
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

func (setter *Setter) getAbout(ctx router.Context) error {
	return setter.h.RespondUrlHTML(ctx, nil)
}

func (setter *Setter) getReading(ctx router.Context) error {
	books, err := goodreads.NewClient(&setter.h.Settings.Goodreads, ctx.Log()).GetBooks()
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
	return setter.h.RespondUrlHTML(ctx, data)
}

func (setter *Setter) getPost(ctx router.Context) error {
	post, err := models.NewPost(ctx.UrlParts()[0])
	if err != nil {
		return err
	}
	return setter.h.RespondHTML(ctx, "post", post)
}

func (setter *Setter) getPosts(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	data := struct {
		Posts []*models.Post
	}{
		posts,
	}
	return setter.h.RespondHTML(ctx, "posts", data)
}

func (setter *Setter) getPostsAtom(ctx router.Context) error {
	posts, err := sortedPosts()
	if err != nil {
		return err
	}

	logoPath := setter.h.Renderer.Webpack().ManifestPath("images/logo.png")
	htmlEntries := atom.PostsToHtmlEntries(posts)
	return setter.h.RespondAtom(ctx, "posts", logoPath, htmlEntries)
}

func (setter *Setter) getRobotsTxt(ctx router.Context) error {
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
