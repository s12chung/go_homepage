package main

import (
	"github.com/sirupsen/logrus"
	"os"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/cmd"
	"github.com/s12chung/go_homepage/go/content/routes"
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

	settings := app.DefaultSettings()
	contentSettings := routes.DefaultSettings()
	settings.Content = contentSettings
	app.ReadFromFile(settings, log)

	err := cmd.NewCmd(settings, func() app.Setter {
		return routes.NewSetter(settings.GeneratedPath, contentSettings, log)
	}).Run(log)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
