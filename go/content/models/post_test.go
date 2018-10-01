package models

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
)

func configFactory() {
	log, _ := logTest.NewNullLogger()
	TestConfig(test.FixturePath, log)
}

func setPostDirEmpty() {
	log, _ := logTest.NewNullLogger()
	TestSetPostDirEmpty(log)
}

func TestMain(m *testing.M) {
	configFactory()
	retCode := m.Run()
	os.Exit(retCode)
}

func TestPost_ID(t *testing.T) {
	post := &Post{Filename: "some_filename"}
	test.AssertLabel(t, "Result", post.ID(), post.Filename)
}

func TestPost_MarkdownFilename(t *testing.T) {
	post := &Post{Filename: "some_filename"}
	test.AssertLabel(t, "Result", post.MarkdownFilename(), post.Filename+".md")
}

func TestPost_FilePath(t *testing.T) {
	testCases := []struct {
		isDraft  bool
		expected string
	}{
		{false, "testdata/posts/some_filename.md"},
		{true, "testdata/drafts/some_filename.md"},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":   testCaseIndex,
			"isDraft": tc.isDraft,
		})
		post := &Post{Filename: "some_filename", IsDraft: tc.isDraft}
		got := post.FilePath()
		if got != tc.expected {
			t.Error(context.GotExpString("Result", got, tc.expected))
		}
	}
}

func TestPost_EditGithubURL(t *testing.T) {
	testCases := []struct {
		githubURL string
		expected  string
	}{
		{"", ""},
		{"https://github.com/s12chung/go_homepage", "https://github.com/s12chung/go_homepage/edit/master/testdata/posts/some_filename.md"},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":     testCaseIndex,
			"githubURL": tc.githubURL,
		})
		factory.settings.GithubURL = tc.githubURL
		post := &Post{Filename: "some_filename"}
		got := post.EditGithubURL()
		if got != tc.expected {
			t.Error(context.GotExpString("Result", got, tc.expected))
		}
	}
}

func Test_NewPost(t *testing.T) {
	testCases := []struct {
		filename string
		cached   bool
	}{
		{"draft1", false},
		{"draft1", true},
		{"draft2", false},
		{"post1", false},
		{"post1", true},
		{"post2", false},
	}

	var prevPost *Post
	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":    testCaseIndex,
			"filename": tc.filename,
			"cached":   tc.cached,
		})

		post, err := NewPost(tc.filename)
		if err != nil {
			t.Error(context.String(err))
		}

		if tc.cached {
			if post != prevPost {
				t.Error(context.GotExpString("matching prevPost", post, prevPost))
			}
		}

		isDraft := strings.Contains(tc.filename, "draft")
		title := strings.Title(tc.filename)

		month := 8
		if isDraft {
			month = 7
		}
		day, err := strconv.ParseInt(title[len(title)-1:], 10, 32)
		if err != nil {
			t.Error(context.String(err))
		}
		exp := &Post{
			title,
			fmt.Sprintf("%v Dec", title),
			time.Date(2017, time.Month(month), int(day), 0, 0, 0, 0, time.UTC),
			tc.filename,
			isDraft,
			fmt.Sprintf("<p>The %v.</p>\n", title),
		}
		if !cmp.Equal(post, exp) {
			t.Error(context.DiffString("Post", post, exp, cmp.Diff(post, exp)))
		}

		prevPost = post
	}
}

func TestAllPosts(t *testing.T) {
	testCases := []struct {
		postDirEmpty bool
		filerType    string
		expLen       int
	}{
		{true, "all", 0},
		{true, "draft", 0},
		{false, "all", 5},
		{false, "empty", 0},
		{false, "draft", 3},
		{false, "noDraft", 2},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"postDirEmpty": tc.postDirEmpty,
			"filerType":    tc.filerType,
		})

		configFactory()
		if tc.postDirEmpty {
			setPostDirEmpty()
		}

		var filter func(*Post) bool
		filterType := tc.filerType
		switch filterType {
		case "all", "empty":
			filter = func(post *Post) bool { return filterType == "all" }
		case "draft", "noDraft":
			filter = func(post *Post) bool { return post.IsDraft == (filterType == "draft") }
		}

		posts, err := AllPosts(filter)
		if err != nil {
			t.Error(context.String(err))
		}
		for _, post := range posts {
			got := filter(post)
			if !got {
				t.Error(context.String("not all posts fix the filter"))
			}
		}
		if len(posts) != tc.expLen {
			t.Error(context.GotExpString("len(posts)", len(posts), tc.expLen))
		}
	}
}

func TestAllPosts_Caching(t *testing.T) {
	post1, err := NewPost("post1")
	if err != nil {
		t.Error(err)
	}

	draft1, err := NewPost("draft1")
	if err != nil {
		t.Error(err)
	}

	posts := sortedAllPosts(t)
	found := 0
	for _, post := range posts {
		if post == post1 || post == draft1 {
			found++
			if found == 2 {
				break
			}
		}
	}
	if found != 2 {
		t.Error("post not using cache from NewPost")
	}

	postsAgain := sortedAllPosts(t)
	for i := range posts {
		if posts[i] != postsAgain[i] {
			t.Errorf("not match post addr with titles: %v, %v", posts[i], postsAgain[i].Title)
		}
	}
}

func sortedAllPosts(t *testing.T) []*Post {
	posts, err := AllPosts(func(post *Post) bool { return true })
	if err != nil {
		t.Error(err)
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Title < posts[j].Title })
	return posts
}

func TestAllPostFilenames(t *testing.T) {
	testCases := []struct {
		postDirEmpty bool
		exp          []string
	}{
		{true, []string{}},
		{false, []string{"post1", "post2", "draft1", "draft2", "draft3"}},
	}

	for _, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"postDirEmpty": tc.postDirEmpty,
		})

		configFactory()
		if tc.postDirEmpty {
			setPostDirEmpty()
		}

		got, err := AllPostFilenames()
		if err != nil {
			t.Error(context.String(err))
		}

		sort.Strings(got)
		sort.Strings(tc.exp)
		if !cmp.Equal(got, tc.exp) {
			t.Error(context.DiffString("Result", got, tc.exp, cmp.Diff(got, tc.exp)))
		}
	}
}
