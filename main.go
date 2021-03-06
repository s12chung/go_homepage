package main

import (
	"os"

	"github.com/s12chung/go_homepage/go/content"
	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/cli"
)

func main() {
	log := app.DefaultLog()

	settings := app.DefaultSettings()
	contentSettings := content.DefaultSettings()
	settings.Content = contentSettings
	app.SettingsFromFile("./settings.json", settings, log)

	theContent := content.NewContent(settings.GeneratedPath, contentSettings, log)
	err := cli.RunDefault(app.NewApp(theContent, settings, log))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
