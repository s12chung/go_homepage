package atom

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/s12chung/go_homepage/go/test"
)

var updateFixturesPtr = test.UpdateFixtureFlag()

func TestRender(t *testing.T) {
	htmlEntries := []*HtmlEntry{
		{"first", "num #1", test.Time(1), "<p>The story starts here</p>", "The sum", test.Time(2)},
		{"second", "num #2", test.Time(3), "<p>The story is in the middle here</p>", "The sum of it all", test.Time(4)},
		{"third", "num #3", test.Time(5), "<p>The story ends here</p>", "The sum of the conclusion", test.Time(6)},
	}

	testCases := []struct {
		htmlEntries []*HtmlEntry
	}{
		{htmlEntries},
		{htmlEntries[0:2]},
		{htmlEntries[0:1]},
		{[]*HtmlEntry{}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index": testCaseIndex,
		})

		bytes, err := Render(DefaultSettings(), "the test feed for posts", "posts", "logo.png", tc.htmlEntries)
		if err != nil {
			t.Error(context.String(err))
		}

		got := string(bytes)
		if len(tc.htmlEntries) == 0 {
			regex := regexp.MustCompile(`<updated>[^<].*</updated>`)
			got = regex.ReplaceAllString(got, "<updated>REPLACED time.Now()</updated>")
		}

		if *updateFixturesPtr {
			test.WriteFixture(t, fmt.Sprintf("feed%v.xml", testCaseIndex), []byte(got))
			continue
		}

		exp := string(test.ReadFixture(t, fmt.Sprintf("feed%v.xml", testCaseIndex)))
		if got != exp {
			t.Error(context.DiffString("Render", got, exp, cmp.Diff(got, exp)))
		}
	}
}
