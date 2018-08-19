package main

import (
	"github.com/sirupsen/logrus"
	"os"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/cmd"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/content/routes"
	"github.com/s12chung/go_homepage/go/lib/html"
)

func main() {
	log := &logrus.Logger{
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

	err := cmd.NewCmd(s, func() app.Setter {
		renderer := html.NewRenderer(s.GeneratedPath, &s.Template, log)
		respondHelper := app.NewRespondHelper(renderer, s)
		return routes.NewSetter(respondHelper)
	}).Run(log)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
