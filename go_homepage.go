package main

import (
	"github.com/sirupsen/logrus"
	"os"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/cli"
	"github.com/s12chung/go_homepage/go/content"
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
	contentSettings := content.DefaultSettings()
	settings.Content = contentSettings
	app.ReadFromFile(settings, log)

	err := cli.NewCli(settings, func() app.Setter {
		return content.NewContent(settings.GeneratedPath, contentSettings, log)
	}).Run(log)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
