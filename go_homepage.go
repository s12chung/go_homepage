package main

import (
	"os"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/cli"
	"github.com/s12chung/go_homepage/go/content"
	"github.com/s12chung/go_homepage/go/lib/utils"
)

func main() {
	log := app.DefaultLog()

	settings := app.DefaultSettings()
	contentSettings := content.DefaultSettings()
	settings.Content = contentSettings
	utils.SettingsFromFile("./settings.json", settings, log)

	application := app.NewApp(content.NewContent(settings.GeneratedPath, contentSettings, log), settings, log)
	err := cli.NewCli(cli.DefaultName(), application).Run(cli.DefaultArgs())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
