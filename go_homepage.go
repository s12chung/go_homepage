package main

import (
	"github.com/sirupsen/logrus"
	"os"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/cli"
	"github.com/s12chung/go_homepage/go/content"
	"github.com/s12chung/go_homepage/go/lib/utils"
)

func main() {
	// when watchman fixes this, change back to Stderr because it is default in logrus:
	// can't downgrade on alpine without manually building...
	// https://github.com/facebook/watchman/issues/602
	log := &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	}

	settings := app.DefaultSettings()
	contentSettings := content.DefaultSettings()
	settings.Content = contentSettings
	utils.SettingsFromFile("./settings.json", settings, log)

	a := app.NewApp(content.NewContent(settings.GeneratedPath, contentSettings, log), settings, log)
	err := cli.NewCli(cli.DefaultName(), a).Run(cli.DefaultArgs())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
