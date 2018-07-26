package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/routes"
	"github.com/s12chung/go_homepage/server"
	"github.com/s12chung/go_homepage/server/router"
	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/utils"
	"github.com/s12chung/go_homepage/view"
)

func main() {
	log := logrus.StandardLogger()

	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	err := NewApp(log).all()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

type App struct {
	Settings settings.Settings
	log      logrus.FieldLogger
}

func NewApp(log logrus.FieldLogger) *App {
	s := *settings.ReadFromFile(log)
	models.Config(s.PostsPath, s.DraftsPath)

	return &App{
		s,
		log,
	}
}

func (app *App) all() error {
	fileServerPtr := flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, generated files in %v on localhost:%v", app.Settings.GeneratedPath, app.Settings.FileServerPort))
	serverPtr := flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", app.Settings.ServerPort))

	flag.Parse()

	if *fileServerPtr {
		return server.RunFileServer(app.Settings.GeneratedPath, app.Settings.FileServerPort, app.log)
	} else if *serverPtr {
		return app.host()
	} else {
		return app.build()
	}
}

func (app *App) host() error {
	var renderer, err = view.NewRenderer(&app.Settings.Template, app.log)
	if err != nil {
		return err
	}

	r := router.NewWebRouter(renderer, &app.Settings, app.log)
	r.FileServe("/assets/", app.Settings.Template.AssetsPath)

	r.Around(func(ctx *router.WebContext, handler func(ctx *router.WebContext) error) error {
		ctx.Log().Infof("Running route: %v", ctx.Url())

		err := handler(ctx)

		if err != nil {
			ctx.Log().Errorf("Error for route %v: %v", ctx.Url(), err)
		} else {
			ctx.Log().Infof("Success for route: %v", ctx.Url())
		}
		return err
	})
	r.GetRootHTML(routes.GetIndex)
	r.GetWildcardHTML(routes.GetPost)
	r.GetHTML("/reading", routes.GetReading)

	return r.Run(app.Settings.ServerPort)
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
	var renderer, err = view.NewRenderer(&app.Settings.Template, app.log)
	if err != nil {
		return err
	}

	var tasks []*pool.Task
	tasks = append(tasks, app.indexPageTask(renderer))
	tasks = append(tasks, app.readingPageTask(renderer))

	eachPostTasks, err := app.eachPostTasks(renderer)
	tasks = append(tasks, eachPostTasks...)

	p := pool.NewPool(tasks, app.Settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		app.log.Errorf("Error found in task - %v", task.Err)
	})

	return nil
}

func (app *App) indexPageTask(renderer *view.Renderer) *pool.Task {
	return pool.NewTask(func() error {
		app.log.Infof("Rendering template: %v", "index")
		bytes, err := renderer.Render("index", nil)
		if err != nil {
			return err
		}
		return writeFile(path.Join(app.Settings.GeneratedPath, "index.html"), bytes)
	})
}

func (app *App) eachPostTasks(renderer *view.Renderer) ([]*pool.Task, error) {
	postsTasks, err := app.eachPostTasksForPath(app.Settings.PostsPath, renderer)
	if err != nil {
		return nil, err
	}

	draftTasks, err := app.eachPostTasksForPath(app.Settings.DraftsPath, renderer)
	if err != nil {
		return nil, err
	}
	return append(postsTasks, draftTasks...), nil
}

func (app *App) eachPostTasksForPath(postsDirPath string, renderer *view.Renderer) ([]*pool.Task, error) {
	filePaths, err := utils.FilePaths(postsDirPath)

	if err != nil {
		if os.IsNotExist(err) {
			app.log.Warnf("Posts path does not exist %v - %v", postsDirPath, err)
			return nil, nil
		}
		return nil, err
	}

	tasks := make([]*pool.Task, len(filePaths))
	for index := range filePaths {
		tasks[index] = pool.NewTask(func() error { return nil })
	}
	return tasks, nil
}

func (app *App) readingPageTask(renderer *view.Renderer) *pool.Task {
	return pool.NewTask(func() error {
		app.log.Infof("Starting task for: %v", "reading")

		bookMap, err := goodreads.NewClient(&app.Settings.Goodreads, app.log).GetBooks()
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
		app.log.Infof("Rendering template: %v", "reading")
		bytes, err := renderer.Render("reading", data)
		if err != nil {
			return err
		}
		return writeFile(path.Join(app.Settings.GeneratedPath, "reading"), bytes)
	})
}

func writeFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}
