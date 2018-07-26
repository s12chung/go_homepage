package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
)

type Post struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	PublishedAt time.Time `yaml:"published_at"`

	Filename     string `yaml:"-"`
	IsDraft      bool   `yaml:"-"`
	MarkdownHTML string `yaml:"-"`
}

func (factory *Factory) NewPost(filename string) (*Post, error) {
	filePath, isDraft, err := factory.postPath(filename)
	if err != nil {
		return nil, err
	}
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	post, markdown, err := postParts(input)
	post.Filename = filename
	post.MarkdownHTML = string(blackfriday.Run([]byte(markdown)))
	post.IsDraft = isDraft

	if err != nil {
		return nil, err
	}
	return post, nil
}

func (factory *Factory) postPath(filename string) (string, bool, error) {
	filename = filename + ".md"
	paths := []string{
		path.Join(factory.postsPath, filename),
		path.Join(factory.draftsPath, filename),
	}

	for index, currentPath := range paths {
		isDraft := index == 1
		_, err := os.Stat(currentPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", isDraft, err
		}
		return currentPath, isDraft, nil
	}
	return "", false, fmt.Errorf("'%v' not found in %v", filename, paths)
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
