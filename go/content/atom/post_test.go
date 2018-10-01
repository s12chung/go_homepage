package atom

import (
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/gostatic/go/test"
	"testing"
)

func TestPostsToHTMLEntries(t *testing.T) {
	testCases := []struct {
		numberOfPosts int
		expected      int
	}{
		{-1, 0},
		{0, 0},
		{1, 1},
		{5, 5},
		{99, 99},
		{100, 100},
		{101, 100},
		{300, 100},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":         testCaseIndex,
			"numberOfPosts": tc.numberOfPosts,
		})

		var posts []*models.Post
		if tc.numberOfPosts != -1 {
			posts = make([]*models.Post, tc.numberOfPosts)
		}
		for i := 0; i < tc.numberOfPosts; i++ {
			posts[i] = &models.Post{}
		}
		entries := PostsToHTMLEntries(posts)
		if len(entries) != tc.expected {
			t.Error(context.GotExpString("len(entries)", len(entries), tc.expected))
		}
	}
}
