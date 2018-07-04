package main

import (
	"time"
	"path"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/view"
)


const generatedDir string = "./generated"
const concurrency int = 10

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	createDir(generatedDir, "")

	var tasks []*pool.Task


	tasks = append(tasks, homepageTasks()...)

	pool.NewPool(tasks, concurrency).LoggedRun()
}

func homepageTasks() ([]*pool.Task) {
	var tasks []*pool.Task
	tasks = append(tasks, pool.NewTask(func() error {
		template := view.NewTemplate("home", path.Join(generatedDir, "index"))

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
		log.Fatal(err)
		os.Exit(1)
		return err
	}

	return nil
}