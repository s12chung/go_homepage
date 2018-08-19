package cmd

import (
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/lib/router"
)

type Cmd struct {
	*app.Settings
	routeSetter func() app.Setter
}

func NewCmd(settings *app.Settings, routeSetter func() app.Setter) *Cmd {
	return &Cmd{settings, routeSetter}
}

func (cmd *Cmd) Run(log logrus.FieldLogger) error {
	fileServerPtr := flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, generated files in %v on localhost:%v", cmd.Settings.GeneratedPath, cmd.Settings.FileServerPort))
	serverPtr := flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", cmd.Settings.ServerPort))

	flag.Parse()

	if *fileServerPtr {
		return router.RunFileServer(cmd.Settings.GeneratedPath, cmd.FileServerPort, log)
	} else {
		a := app.NewApp(cmd.routeSetter(), cmd.Settings, log)
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
