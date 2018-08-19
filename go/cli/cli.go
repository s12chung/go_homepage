package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/lib/router"
)

type Cli struct {
	*app.Settings
	routeSetter func() app.Setter
}

func NewCli(settings *app.Settings, routeSetter func() app.Setter) *Cli {
	return &Cli{settings, routeSetter}
}

func (cli *Cli) Run(log logrus.FieldLogger) error {
	fileServerPtr := flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, generated files in %v on localhost:%v", cli.Settings.GeneratedPath, cli.Settings.FileServerPort))
	serverPtr := flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", cli.Settings.ServerPort))

	flag.Parse()

	if *fileServerPtr {
		return router.RunFileServer(cli.Settings.GeneratedPath, cli.FileServerPort, log)
	} else {
		a := app.NewApp(cli.routeSetter(), cli.Settings, log)
		if *serverPtr {
			return a.Host()
		} else {
			start := time.Now()
			defer func() {
				log.Infof("Build generated in %v.", time.Now().Sub(start))
			}()
			return a.Generate()
		}
	}
}
