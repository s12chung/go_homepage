package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/models"
	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/routes"
	"github.com/s12chung/go_homepage/server"
	"github.com/s12chung/go_homepage/server/router"
	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view"
)

func main() {
	var log logrus.FieldLogger
	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	log = &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	}

	s := settings.ReadFromFile(log)
	models.Config(s.PostsPath, s.DraftsPath, log.WithFields(logrus.Fields{
		"type": "models",
	}))

	err := NewApp(s, log).all()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

type App struct {
	Settings *settings.Settings
	log      logrus.FieldLogger
}

func NewApp(settings *settings.Settings, log logrus.FieldLogger) *App {
	return &App{
		settings,
		log,
	}
}

func (app *App) all() error {
	fileServerPtr := flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, generated files in %v on localhost:%v", app.Settings.GeneratedPath, app.Settings.FileServerPort))
	serverPtr := flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", app.Settings.ServerPort))

	flag.Parse()

	if *fileServerPtr {
		return server.RunFileServer(app.Settings.GeneratedPath, app.Settings.FileServerPort, app.log)
	} else {
		if *serverPtr {
			return app.host()
		} else {
			return app.build()
		}
	}
}

func (app *App) host() error {
	var renderer, err = view.NewRenderer(&app.Settings.Template, app.log)
	if err != nil {
		return err
	}

	r := router.NewWebRouter(renderer, app.Settings, app.log)
	r.FileServe("/assets/", app.Settings.Template.AssetsPath)
	setRoutes(r)

	return r.Run(app.Settings.ServerPort)
}

func (app *App) build() error {
	err := app.setup()
	if err != nil {
		return err
	}

	renderer, err := view.NewRenderer(&app.Settings.Template, app.log)
	if err != nil {
		return err
	}

	r := router.NewGenerateRouter(renderer, app.Settings, app.log)
	setRoutes(r)
	err = app.requestRoutes(r)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) setup() error {
	return os.MkdirAll(app.Settings.GeneratedPath, 0755)
}

func setRoutes(r router.Router) {
	r.Around(func(ctx router.Context, handler func(ctx router.Context) error) error {
		ctx.SetLog(ctx.Log().WithFields(logrus.Fields{
			"type": "routes",
			"url":  ctx.Url(),
		}))

		var err error

		ctx.Log().Infof("Running route")
		start := time.Now()
		defer func() {
			ending := fmt.Sprintf(" for route")

			log := ctx.Log().WithField("time", time.Now().Sub(start))
			if err != nil {
				log.Errorf("Error"+ending+" - %v", err)
			} else {
				log.Infof("Success" + ending)
			}
		}()

		err = handler(ctx)
		return err
	})
	routes.SetRoutes(r)
}

func (app *App) requestRoutes(r router.Router) error {
	requester := r.Requester()

	allUrls, err := app.allUrls(r)
	if err != nil {
		return err
	}

	tasks := make([]*pool.Task, len(allUrls)-len(routes.DependentUrls))
	i := 0
	for _, url := range allUrls {
		_, exists := routes.DependentUrls[url]
		if !exists {
			tasks[i] = app.getUrlTask(requester, url)
			i += 1
		}
	}
	app.runTasks(tasks)

	tasks = make([]*pool.Task, len(routes.DependentUrls))
	i = 0
	for url := range routes.DependentUrls {
		tasks[i] = app.getUrlTask(requester, url)
		i += 1
	}
	app.runTasks(tasks)

	return nil
}

func (app *App) getUrlTask(requester router.Requester, url string) *pool.Task {
	log := app.log.WithFields(logrus.Fields{
		"type": "task",
		"url":  url,
	})

	return pool.NewTask(log, func() error {
		bytes, err := requester.Get(url)
		if err != nil {
			return err
		}

		filename := url
		if url == router.RootUrlPattern {
			filename = "index.html"
		}

		generatedFilePath := path.Join(app.Settings.GeneratedPath, filename)

		log.Infof("Writing response into %v", generatedFilePath)
		return writeFile(generatedFilePath, bytes)
	})
}

func writeFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}

func (app *App) allUrls(r router.Router) ([]string, error) {
	allUrls := r.StaticRoutes()
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

func (app *App) runTasks(tasks []*pool.Task) {
	p := pool.NewPool(tasks, app.Settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		task.Log.Errorf("Error for task - %v", task.Error)
	})
}
