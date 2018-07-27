package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

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
	} else {
		var renderer, err = view.NewRenderer(&app.Settings.Template, app.log)
		if err != nil {
			return err
		}
		if *serverPtr {
			return app.host(renderer)
		} else {
			return app.build(renderer)
		}
	}
}

func (app *App) host(renderer *view.Renderer) error {
	r := router.NewWebRouter(renderer, &app.Settings, app.log)
	r.FileServe("/assets/", app.Settings.Template.AssetsPath)
	setRoutes(r)

	return r.Run(app.Settings.ServerPort)
}

func (app *App) build(renderer *view.Renderer) error {
	var err error
	if err = app.setup(); err != nil {
		return err
	}

	r := router.NewGenerateRouter(renderer, &app.Settings, app.log)
	setRoutes(r)
	if err = app.requestRoutes(r); err != nil {
		return err
	}
	return nil
}

func (app *App) setup() error {
	return os.MkdirAll(app.Settings.GeneratedPath, 0755)
}

func setRoutes(r router.Router) {
	r.Around(func(ctx router.Context, handler func(ctx router.Context) error) error {
		var err error

		start := time.Now()
		defer func() {
			ending := fmt.Sprintf(" for route %v (%v)", ctx.Url(), time.Now().Sub(start))
			if err != nil {
				ctx.Log().Errorf("Error"+ending+" - %v", err)
			} else {
				ctx.Log().Infof("Success" + ending)
			}
		}()

		ctx.Log().Infof("Running route: %v", ctx.Url())

		err = handler(ctx)
		return err
	})
	routes.SetRoutes(r)
}

func (app *App) requestRoutes(r router.Router) error {
	allUrls, err := app.allUrls(r)
	if err != nil {
		return err
	}

	requester := r.Requester()

	tasks := make([]*pool.Task, len(allUrls))
	for i, url := range allUrls {
		tasks[i] = app.getUrlTask(requester, url)
	}

	p := pool.NewPool(tasks, app.Settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		app.log.Errorf("Error found in task - %v", task.Err)
	})

	return nil
}

func (app *App) getUrlTask(requester router.Requester, url string) *pool.Task {
	return pool.NewTask(func() error {
		bytes, err := requester.Get(url)
		if err != nil {
			return err
		}

		filename := url
		if url == router.RootUrlPattern {
			filename = "index.html"
		}

		generatedFilePath := path.Join(app.Settings.GeneratedPath, filename)
		app.log.Infof("Writing response for URL %v into FILE_PATH %v", url, generatedFilePath)
		return writeFile(generatedFilePath, bytes)
	})
}

func writeFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}

func (app *App) allUrls(r router.Router) ([]string, error) {
	allUrls := r.StaticRoutes()

	postsUrls, err := app.eachPostUrl(app.Settings.PostsPath)
	if err != nil {
		return nil, err
	}
	allUrls = append(allUrls, postsUrls...)

	draftUrls, err := app.eachPostUrl(app.Settings.DraftsPath)
	if err != nil {
		return nil, err
	}
	allUrls = append(allUrls, draftUrls...)

	return allUrls, nil
}

func (app *App) eachPostUrl(postsDirPath string) ([]string, error) {
	filePaths, err := utils.FilePaths(postsDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			app.log.Warnf("Posts path does not exist %v - %v", postsDirPath, err)
			return nil, nil
		}
		return nil, err
	}

	urls := make([]string, len(filePaths))
	for i, filePath := range filePaths {
		basename := filepath.Base(filePath)
		urls[i] = fmt.Sprintf("/%v", strings.TrimSuffix(basename, filepath.Ext(basename)))
	}
	return urls, nil
}
