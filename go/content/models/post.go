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

	"github.com/s12chung/gostatic/go/lib/utils"
)

const markdownExtension = ".md"

var postMap = map[string]*Post{}

func ResetPostMap() { postMap = map[string]*Post{} }

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

func (post *Post) MarkdownFilename() string {
	return markdownFilename(post.Filename)
}

func (post *Post) FilePath() string {
	folderPath := factory.settings.PostsPath
	if post.IsDraft {
		folderPath = factory.settings.DraftsPath
	}
	return strings.Join([]string{
		utils.CleanFilePath(folderPath),
		post.MarkdownFilename(),
	}, "/")
}

func (post *Post) EditGithubUrl() string {
	githubUrl := factory.settings.GithubUrl
	if githubUrl == "" {
		return ""
	}

	return strings.Join([]string{
		githubUrl,
		"edit/master",
		post.FilePath(),
	}, "/")
}

func NewPost(filename string) (*Post, error) {
	post := postMap[filename]
	if post != nil {
		return post, nil
	}

	filePath, isDraft, err := postPath(filename)
	if err != nil {
		return nil, err
	}
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	post, markdown, err := postParts(input)
	if err != nil {
		return nil, err
	}
	post.Filename = filename
	post.MarkdownHTML = string(blackfriday.Run([]byte(markdown)))
	post.IsDraft = isDraft
	postMap[post.Filename] = post
	return post, nil
}

func Posts() ([]*Post, error) {
	return AllPosts(func(post *Post) bool { return !post.IsDraft })
}

func AllPosts(sel func(*Post) bool) ([]*Post, error) {
	err := fillPostMap()
	if err != nil {
		return nil, err
	}
	return toPosts(postMap, sel), nil
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
	allPostUrls := []string{}

	postsUrls, err := postFilenames(factory.settings.PostsPath)
	if err != nil {
		return nil, err
	}
	allPostUrls = append(allPostUrls, postsUrls...)

	draftUrls, err := postFilenames(factory.settings.DraftsPath)
	if err != nil {
		return nil, err
	}
	return append(allPostUrls, draftUrls...), nil
}

func postFilenames(postsDirPath string) ([]string, error) {
	filePaths, err := utils.FilePaths(markdownExtension, postsDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			factory.log.Warnf("Posts path does not exist %v - %v", postsDirPath, err)
			return nil, nil
		}
		return nil, err
	}

	filenames := make([]string, len(filePaths))
	for i, filePath := range filePaths {
		basename := filepath.Base(filePath)
		filenames[i] = strings.TrimSuffix(basename, filepath.Ext(basename))
	}
	return filenames, nil
}

func postPath(filename string) (string, bool, error) {
	filename = markdownFilename(filename)
	paths := []string{
		path.Join(factory.settings.PostsPath, filename),
		path.Join(factory.settings.DraftsPath, filename),
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

func markdownFilename(filename string) string {
	return filename + ".md"
}
