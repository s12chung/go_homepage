package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"

	"github.com/s12chung/go_homepage/goodreads"
	"github.com/s12chung/go_homepage/pool"
	"github.com/s12chung/go_homepage/server"
	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/utils"
	"github.com/s12chung/go_homepage/view"
)

func main() {
	start := time.Now()
	defer func() {
		log.Infof("Build completed in %v.", time.Now().Sub(start))
	}()

	err := NewApp().all()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

type App struct {
	Settings settings.Settings
}

func NewApp() *App {
	return &App{*settings.ReadFromFile()}
}

func (app *App) all() error {
	runServerPtr := flag.Bool("server", false, "Serves page on localhost:3000")

	flag.Parse()

	if *runServerPtr {
		return server.Run(app.Settings.GeneratedPath, app.Settings.ServerPort)
	} else {
		return app.build()
	}
}

func (app *App) build() error {
	var err error
	if err = app.setup(); err != nil {
		return err
	}
	if err = app.runTasks(); err != nil {
		return err
	}
	return nil
}

func (app *App) setup() error {
	return os.MkdirAll(app.Settings.GeneratedPath, 0755)
}

func (app *App) runTasks() error {
	var templateGenerator, err = view.NewTemplateGenerator(app.Settings.Template)
	if err != nil {
		return err
	}

	var tasks []*pool.Task
	tasks = append(tasks, app.indexPageTask(templateGenerator))
	tasks = append(tasks, app.readingPageTask(templateGenerator))

	eachPostTasks, err := app.eachPostTasks(templateGenerator)
	tasks = append(tasks, eachPostTasks...)

	p := pool.NewPool(tasks, app.Settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		log.Errorf("Error found in task - %v", task.Err)
	})

	return nil
}

func (app *App) indexPageTask(templateGenerator *view.TemplateGenerator) *pool.Task {
	return pool.NewTask(func() error {
		log.Infof("Rendering template: %v", "index")
		bytes, err := templateGenerator.NewTemplate("index").Render(nil)
		if err != nil {
			return err
		}
		return writeFile(path.Join(app.Settings.GeneratedPath, "index.html"), bytes)
	})
}

func (app *App) eachPostTasks(templateGenerator *view.TemplateGenerator) ([]*pool.Task, error) {
	postsTasks, err := app.eachPostTasksForPath(app.Settings.PostsPath, templateGenerator)
	if err != nil {
		return nil, err
	}

	draftTasks, err := app.eachPostTasksForPath(app.Settings.DraftsPath, templateGenerator)
	if err != nil {
		return nil, err
	}
	return append(postsTasks, draftTasks...), nil
}

func (app *App) eachPostTasksForPath(postsDirPath string, templateGenerator *view.TemplateGenerator) ([]*pool.Task, error) {
	filePaths, err := utils.FilePaths(postsDirPath)

	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("Posts path does not exist %v - %v", postsDirPath, err)
			return nil, nil
		}
		return nil, err
	}

	tasks := make([]*pool.Task, len(filePaths))
	for index, filePath := range filePaths {
		currentPath := filePath
		tasks[index] = pool.NewTask(func() error {
			log.Infof("Starting task for: %v - %v", "post", currentPath)

			input, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}
			post, markdown, err := postParts(input)

			data := struct {
				Post         Post
				MarkdownHTML template.HTML
			}{
				*post,
				template.HTML(blackfriday.Run([]byte(markdown))),
			}
			log.Infof("Rendering template: %v - %v", "post", currentPath)

			bytes, err := templateGenerator.NewTemplate("post").Render(data)
			if err != nil {
				return err
			}
			return writeFile(path.Join(app.Settings.GeneratedPath, "post"), bytes)
		})
	}
	return tasks, nil
}

type Post struct {
	Title       string    `yaml:"title"`
	PublishedAt time.Time `yaml:"published_at"`
}

func postParts(bytes []byte) (*Post, string, error) {
	frontMatter, markdown, err := splitFrontMatter(string(bytes))
	if err != nil {
		return nil, "", err
	}

	post := Post{}
	yaml.Unmarshal([]byte(frontMatter), &post)
	return &post, markdown, nil
}

func splitFrontMatter(content string) (string, string, error) {
	parts := regexp.MustCompile("(?m)^---").Split(content, 3)

	if len(parts) == 3 && parts[0] == "" {
		return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), nil
	}

	return "", "", fmt.Errorf("FrontMatter format is not followed")
}

func (app *App) readingPageTask(templateGenerator *view.TemplateGenerator) *pool.Task {
	return pool.NewTask(func() error {
		log.Infof("Starting task for: %v", "reading")

		bookMap, err := goodreads.NewClient(app.Settings.Goodreads).GetBooks()
		if err != nil {
			return err
		}

		books := goodreads.ToBooks(bookMap)
		sort.Slice(books, func(i, j int) bool { return books[i].SortedDate().After(books[j].SortedDate()) })

		data := struct {
			Books        []goodreads.Book
			RatingMap    map[int]int
			EarliestYear int
			Today        time.Time
		}{
			books,
			goodreads.RatingMap(bookMap),
			books[len(books)-1].SortedDate().Year(),
			time.Now(),
		}
		log.Infof("Rendering template: %v", "reading")
		bytes, err := templateGenerator.NewTemplate("reading").Render(data)
		if err != nil {
			return err
		}
		return writeFile(path.Join(app.Settings.GeneratedPath, "reading"), bytes)
	})
}

func writeFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}
