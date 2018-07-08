package main

import (
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"

	"flag"
	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/server"
	"github.com/s12chung/go_homepage/view"
)

const generatedPath = "./generated"
const generatedAssetsPath = "./generated/assets"
const concurrency = 10

const serverPort = 3000

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	err := all()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func all() error {
	runServerPtr := flag.Bool("server", false, "Serves page on a willServe")
	flag.Parse()

	if *runServerPtr {
		return server.Run(generatedPath, serverPort)
	} else {
		return build()
	}
}

func build() error {
	var err error
	if err = setup(); err != nil {
		return err
	}
	if err = runTasks(); err != nil {
		return err
	}
	return nil
}

func setup() error {
	return os.MkdirAll(generatedPath, 0755)
}

func runTasks() error {
	var templateGenerator, err = view.NewTemplateGenerator(generatedAssetsPath)
	if err != nil {
		return nil
	}

	var tasks []*pool.Task

	tasks = append(tasks, homepageTasks(templateGenerator)...)

	pool.NewPool(tasks, concurrency).LoggedRun()

	return nil
}

func homepageTasks(templateGenerator *view.TemplateGenerator) []*pool.Task {
	var tasks []*pool.Task
	tasks = append(tasks, pool.NewTask(func() error {
		template := templateGenerator.NewTemplate("home", path.Join(generatedPath, "index"))

		err := template.Render()
		if err != nil {
			log.Fatal(err)
			return err
		}

		return nil
	}))
	return tasks
}
