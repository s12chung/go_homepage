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
	apiGetPtr := flag.Bool("api-get", false, "Only gets api data")

	flag.Parse()

	if *apiGetPtr {
		return app.apiGet()
	} else if *runServerPtr {
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
	err := os.MkdirAll(app.Settings.GeneratedPath, 0755)
	if err != nil {
		return err
	}
	return app.apiGet()
}

func (app *App) apiGet() error {
	return goodreads.NewClient(app.Settings.Goodreads).Get()
}

func (app *App) runTasks() error {
	var templateGenerator, err = view.NewTemplateGenerator(app.Settings.GeneratedPath, app.Settings.Template)
	if err != nil {
		return nil
	}

	var tasks []*pool.Task

	tasks = append(tasks, generateTasks(templateGenerator, []string{"index"})...)

	pool.NewPool(tasks, app.Settings.Concurrency).LoggedRun()

	return nil
}

func generateTasks(templateGenerator *view.TemplateGenerator, names []string) []*pool.Task {
	var tasks []*pool.Task
	for _, name := range names {
		tasks = append(tasks, pool.NewTask(func() error {
			return templateGenerator.RenderNewTemplate(name)
		}))
	}
	return tasks
}
