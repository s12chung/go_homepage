package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"flag"
	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/server"
	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view"
)

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	err := NewApp().all()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

type App struct {
	Settings settings.Settings
}

func NewApp() *App {
	return &App{*settings.ReadFromFile()}
}

func (app *App) all() error {
	runServerPtr := flag.Bool("server", false, "Serves page on localhost:3000")

	flag.Parse()

	if *runServerPtr {
		return server.Run(app.Settings.GeneratedPath, app.Settings.ServerPort)
	} else {
		return app.build()
	}
}

func (app *App) build() error {
	var err error
	if err = app.setup(); err != nil {
		return err
	}
	if err = app.runTasks(); err != nil {
		return err
	}
	return nil
}

func (app *App) setup() error {
	return os.MkdirAll(app.Settings.GeneratedPath, 0755)
}

func (app *App) runTasks() error {
	var templateGenerator, err = view.NewTemplateGenerator(app.Settings.GeneratedPath, app.Settings.Template)
	if err != nil {
		return nil
	}

	var tasks []*pool.Task
	tasks = append(tasks, pool.NewTask(func() error {
		return templateGenerator.RenderNewTemplate("index", nil)
	}))
	tasks = append(tasks, app.readingPageTask(templateGenerator))

	p := pool.NewPool(tasks, app.Settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		log.Errorf("Error found in task - %v", task.Err)
	})

	return nil
}

func (app *App) readingPageTask(templateGenerator *view.TemplateGenerator) *pool.Task {
	return pool.NewTask(func() error {
		bookMap, err := goodreads.NewClient(app.Settings.Goodreads).GetBooks()
		if err != nil {
			return err
		}

		books := make([]goodreads.Book, len(bookMap))
		i := 0
		for _, v := range bookMap {
			books[i] = v
			i += 1
		}
		data := struct {
			Books []goodreads.Book
		}{books}
		return templateGenerator.RenderNewTemplate("reading", data)
	})
}
