package routes

import (
	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/server/router"
	"sort"
	"time"
)

func GetIndex(ctx *router.WebContext) error {
	return ctx.Render("index", nil)
}

func GetReading(ctx *router.WebContext) error {
	bookMap, err := goodreads.NewClient(&ctx.Settings().Goodreads, ctx.Log()).GetBooks()
	if err != nil {
		return err
	}

	books := goodreads.ToBooks(bookMap)
	sort.Slice(books, func(i, j int) bool { return books[i].SortedDate().After(books[j].SortedDate()) })

	data := struct {
		Books        []goodreads.Book
		RatingMap    map[int]int
		EarliestYear int
		Today        time.Time
	}{
		books,
		goodreads.RatingMap(bookMap),
		books[len(books)-1].SortedDate().Year(),
		time.Now(),
	}
	return ctx.Render("reading", data)
}

func GetPost(ctx *router.WebContext) error {
	filename := ctx.UrlParts()[0]
	post, err := models.F.NewPost(filename)
	if err != nil {
		return err
	}
	return ctx.Render("post", post)
}
