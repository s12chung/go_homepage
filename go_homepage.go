package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"flag"
	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/server"
	"github.com/s12chung/go_homepage/view"
)

const generatedPath = "./generated"
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
	runServerPtr := flag.Bool("server", false, "Serves page on localhost:3000")
	apiGetPtr := flag.Bool("api-get", false, "Only gets api data")

	flag.Parse()

	if *apiGetPtr {
		return apiGet()
	} else if *runServerPtr {
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
	err := os.MkdirAll(generatedPath, 0755)
	if err != nil {
		return err
	}
	return apiGet()
}

func apiGet() error {
	return goodreads.Get()
}

func runTasks() error {
	var templateGenerator, err = view.NewTemplateGenerator(generatedPath)
	if err != nil {
		return nil
	}

	var tasks []*pool.Task

	tasks = append(tasks, generateTasks(templateGenerator, []string{"index"})...)

	pool.NewPool(tasks, concurrency).LoggedRun()

	return nil
}

func generateTasks(templateGenerator *view.TemplateGenerator, names []string) []*pool.Task {
	var tasks []*pool.Task
	for _, name := range names {
		tasks = append(tasks, pool.NewTask(func() error {
			return templateGenerator.RenderNewTemplate(name)
		}))
	}
	return tasks
}
