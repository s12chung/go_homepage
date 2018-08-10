package main

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/app/respond"
	"github.com/s12chung/go_homepage/go/app/settings"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/content/routes"
	"github.com/s12chung/go_homepage/go/lib/html"
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
	models.Config(s.PostsPath, s.DraftsPath, s.GithubUrl, log.WithFields(logrus.Fields{
		"type": "models",
	}))

	renderer := html.NewRenderer(s.GeneratedPath, &s.Template, log)
	respondHelper := respond.NewHelper(renderer, s)
	routeSetter := routes.NewSetter(respondHelper)

	err := app.NewApp(routeSetter, s, log).Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
