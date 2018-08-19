package cli

import (
	"flag"
	"fmt"
	"os"
)

//go:generate mockgen -destination=../test/mocks/cli_app.go -package=mocks github.com/s12chung/go_homepage/go/cli App
type App interface {
	RunFileServer() error
	Host() error
	Generate() error

	GeneratedPath() string
	FileServerPort() int
	ServerPort() int
}

func DefaultName() string {
	return os.Args[0]
}

func DefaultArgs() []string {
	return os.Args[1:]
}

type Cli struct {
	app  App
	flag *flag.FlagSet
}

func NewCli(name string, app App) *Cli {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	return &Cli{app, f}
}

func (cli *Cli) Run(args []string) error {
	app := cli.app

	fileServerPtr := cli.flag.Bool("file-server", false, fmt.Sprintf("Serves, but not generates, files in %v on localhost:%v", app.GeneratedPath(), app.FileServerPort()))
	serverPtr := cli.flag.Bool("server", false, fmt.Sprintf("Hosts server on localhost:%v", app.ServerPort()))
	err := cli.flag.Parse(args)
	if err != nil {
		return nil
	}

	if *fileServerPtr {
		return app.RunFileServer()
	} else {
		if *serverPtr {
			return app.Host()
		} else {
			return app.Generate()
		}
	}
}
