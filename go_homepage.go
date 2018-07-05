package main

import (
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/view"
)

const generatedPath = "./generated"
const assetsPath = "./generated/assets"
const concurrency = 10

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	err := setup()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	setup()
	runTasks()
}

func setup() error {
	if err := createDir(generatedPath, ""); err != nil {
		return err
	}
	return nil
}

func runTasks() error {
	var templateGenerator, err = view.NewTemplateGenerator(assetsPath)
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

func createDir(targetDir, newDir string) error {
	dir := path.Join(targetDir, newDir)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return nil
}
