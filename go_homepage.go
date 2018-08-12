package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/content/routes"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/router"
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

	s := app.ReadFromFile(log)
	models.Config(s.PostsPath, s.DraftsPath, s.GithubUrl, log.WithFields(logrus.Fields{
		"type": "models",
	}))

	err := run(s, log)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run(s *app.Settings, log logrus.FieldLogger) error {
	fileServerPtr := flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, generated files in %v on localhost:%v", s.GeneratedPath, s.FileServerPort))
	serverPtr := flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", s.ServerPort))

	flag.Parse()

	if *fileServerPtr {
		return router.RunFileServer(s.GeneratedPath, s.FileServerPort, log)
	} else {
		renderer := html.NewRenderer(s.GeneratedPath, &s.Template, log)
		respondHelper := app.NewRespondHelper(renderer, s)
		routeSetter := routes.NewSetter(respondHelper)
		a := app.NewApp(routeSetter, s, log)

		if *serverPtr {
			return a.Host()
		} else {
			return a.Generate()
		}
	}
}
