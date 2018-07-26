package routes

import (
	"sort"
	"time"

	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/server"
)

func GetIndex(ctx *server.Context) error {
	ctx.Log.Infof("Rendering template: %v", "index")
	bytes, err := ctx.Renderer.Render("index", nil)
	if err != nil {
		return err
	}
	return ctx.Write(bytes)
}

func GetReading(ctx *server.Context) error {
	ctx.Log.Infof("Starting task for: %v", "reading")

	bookMap, err := goodreads.NewClient(&ctx.Settings.Goodreads, ctx.Log).GetBooks()
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
	ctx.Log.Infof("Rendering template: %v", "reading")
	bytes, err := ctx.Renderer.Render("reading", data)
	if err != nil {
		return err
	}

	return ctx.Write(bytes)
}
