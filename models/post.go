package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"

	"github.com/s12chung/go_homepage/utils"
)

var postMap = map[string]*Post{}

type Post struct {
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	PublishedAt time.Time `yaml:"published_at"`

	Filename     string `yaml:"-"`
	IsDraft      bool   `yaml:"-"`
	MarkdownHTML string `yaml:"-"`
}

func (post *Post) Id() string {
	return post.Filename
}

func NewPost(filename string) (*Post, error) {
	return factory.NewPost(filename)
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
	postMap[post.Filename] = post
	return post, nil
}

func AllPosts(sel func(*Post) bool) ([]*Post, error) {
	err := fillPostMap()
	if err != nil {
		return nil, err
	}
	return toPosts(postMap, sel), nil
}

func Posts() ([]*Post, error) {
	return AllPosts(func(post *Post) bool { return !post.IsDraft })
}

func Drafts() ([]*Post, error) {
	return AllPosts(func(post *Post) bool { return post.IsDraft })
}

func fillPostMap() error {
	allPostFilenames, err := AllPostFilenames()
	if err != nil {
		return err
	}

	if len(allPostFilenames) != len(postMap) {
		for _, filename := range allPostFilenames {
			_, exists := postMap[filename]
			if !exists {
				_, err := NewPost(filename)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func toPosts(postMap map[string]*Post, sel func(*Post) bool) []*Post {
	if sel == nil {
		sel = func(post *Post) bool { return true }
	}

	var posts []*Post
	for _, post := range postMap {
		if sel(post) {
			posts = append(posts, post)
		}
	}
	return posts
}

func AllPostFilenames() ([]string, error) {
	return factory.allPostFilenames()
}

func (factory *Factory) allPostFilenames() ([]string, error) {
	var allPostUrls []string

	postsUrls, err := factory.postFilenames(factory.postsPath)
	if err != nil {
		return nil, err
	}
	allPostUrls = append(allPostUrls, postsUrls...)

	draftUrls, err := factory.postFilenames(factory.draftsPath)
	if err != nil {
		return nil, err
	}
	return append(allPostUrls, draftUrls...), nil
}

func (factory *Factory) postFilenames(postsDirPath string) ([]string, error) {
	filePaths, err := utils.FilePaths(".md", postsDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			factory.log.Warnf("Posts path does not exist %v - %v", postsDirPath, err)
			return nil, nil
		}
		return nil, err
	}

	urls := make([]string, len(filePaths))
	for i, filePath := range filePaths {
		basename := filepath.Base(filePath)
		urls[i] = strings.TrimSuffix(basename, filepath.Ext(basename))
	}
	return urls, nil
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
