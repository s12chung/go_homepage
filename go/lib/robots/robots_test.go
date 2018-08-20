package robots

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/s12chung/go_homepage/go/test"
)

var updateFixturesPtr = test.UpdateFixtureFlag()

func TestToFileString(t *testing.T) {
	all := NewUserAgent(EverythingUserAgent, []string{"/"})
	twitterForPages := NewUserAgent("Applebot", []string{"/ajax/", "/blah"})
	baiduForPages := NewUserAgent("baiduspider", []string{"/assets/", "/favicon.ico"})
	naverForAll := NewUserAgent("Naverbot", []string{"/"})

	testCases := []struct {
		userAgents []*UserAgent
	}{
		{[]*UserAgent{all}},
		{[]*UserAgent{all, twitterForPages}},
		{[]*UserAgent{twitterForPages, baiduForPages, naverForAll}},
		{[]*UserAgent{all, twitterForPages, baiduForPages, naverForAll}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index": testCaseIndex,
		})

		got := ToFileString(tc.userAgents)
		fixtureFilename := fmt.Sprintf("robots%v.txt", testCaseIndex)
		if *updateFixturesPtr {
			test.WriteFixture(t, fixtureFilename, []byte(got))
			continue
		}

		exp := string(test.ReadFixture(t, fixtureFilename))
		if got != exp {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}
