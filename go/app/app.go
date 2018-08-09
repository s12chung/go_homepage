package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app/routes"
	"github.com/s12chung/go_homepage/go/app/settings"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/pool"
	"github.com/s12chung/go_homepage/go/lib/router"
)

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

func (app *App) Run() error {
	fileServerPtr := flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, generated files in %v on localhost:%v", app.Settings.GeneratedPath, app.Settings.FileServerPort))
	serverPtr := flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", app.Settings.ServerPort))

	flag.Parse()

	if *fileServerPtr {
		return router.RunFileServer(app.Settings.GeneratedPath, app.Settings.FileServerPort, app.log)
	} else {
		if *serverPtr {
			return app.host()
		} else {
			return app.build()
		}
	}
}

func (app *App) host() error {
	var renderer = html.NewRenderer(app.Settings.GeneratedPath, &app.Settings.Template, app.log)
	r := router.NewWebRouter(app.Settings.ServerPort, app.log)
	r.FileServe(renderer.AssetsUrl(), renderer.GenereratedAssetsPath())
	app.setRoutes(r, renderer)

	return r.Run()
}

func (app *App) build() error {
	err := app.setup()
	if err != nil {
		return err
	}

	renderer := html.NewRenderer(app.Settings.GeneratedPath, &app.Settings.Template, app.log)
	r := router.NewGenerateRouter(app.log)
	routeSetter := app.setRoutes(r, renderer)
	err = app.requestRoutes(routeSetter)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) setup() error {
	return os.MkdirAll(app.Settings.GeneratedPath, 0755)
}

func (app *App) setRoutes(r router.Router, renderer *html.Renderer) *routes.RouteSetter {
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

	routeSetter := routes.NewRouteSetter(r, renderer, app.Settings)
	routeSetter.SetRoutes()
	return routeSetter
}

func (app *App) requestRoutes(routeSetter *routes.RouteSetter) error {
	requester := routeSetter.Router.Requester()

	var urlBatches [][]string

	independentUrls, err := routeSetter.IndependentUrls()
	if err != nil {
		return err
	}

	urlBatches = append(urlBatches, independentUrls)
	urlBatches = append(urlBatches, routeSetter.DependentUrls())

	for _, urlBatch := range urlBatches {
		app.runTasks(app.urlsToTasks(requester, urlBatch))
	}
	return nil
}

func (app *App) urlsToTasks(requester router.Requester, urls []string) []*pool.Task {
	tasks := make([]*pool.Task, len(urls))
	for i, url := range urls {
		tasks[i] = app.getUrlTask(requester, url)
	}
	return tasks
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

func (app *App) runTasks(tasks []*pool.Task) {
	p := pool.NewPool(tasks, app.Settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		task.Log.Errorf("Error for task - %v", task.Error)
	})
}
