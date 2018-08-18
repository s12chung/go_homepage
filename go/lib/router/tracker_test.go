package router

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
	"sort"
	"testing"
)

func defaultTracker(allUrls func() ([]string, error)) *Tracker {
	return NewTracker(allUrls)
}

func TestTracker_Urls(t *testing.T) {
	testCases := []struct {
		allUrls         []string
		allUrlsError    bool
		dependentUrls   []string
		independentUrls []string
	}{
		{[]string{}, false, []string{}, []string{}},
		{[]string{}, true, []string{}, []string{}},
		{[]string{}, true, []string{"a", "b"}, []string{}},
		{[]string{"a", "b"}, false, []string{"a", "b"}, []string{}},
		{[]string{"a", "b", "c", "d"}, false, []string{"a", "b"}, []string{"c", "d"}},
		{[]string{"a", "b", "c", "d"}, false, []string{}, []string{"a", "b", "c", "d"}},
		{[]string{"a", "b"}, false, []string{}, []string{"a", "b"}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":         testCaseIndex,
			"allUrls":       tc.allUrls,
			"allUrlsError":  tc.allUrlsError,
			"dependentUrls": tc.dependentUrls,
		})

		tracker := defaultTracker(func() ([]string, error) {
			if tc.allUrlsError {
				return nil, fmt.Errorf("error")
			}
			return tc.allUrls, nil
		})
		for _, dependentUrl := range tc.dependentUrls {
			tracker.AddDependentUrl(dependentUrl)
		}

		got := tracker.DependentUrls()
		exp := tc.dependentUrls
		sort.Strings(got)
		sort.Strings(exp)

		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("dependentUrls", got, exp, cmp.Diff(got, exp)))
		}
		got, err := tracker.IndependentUrls()
		if err != nil {
			if tc.allUrlsError == false {
				t.Error(context.String(err))
			}
			if got != nil {
				t.Error(context.String("independentUrls should be nil with error"))
			}
			continue
		}
		exp = tc.independentUrls
		sort.Strings(got)
		sort.Strings(exp)
		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("independentUrls", got, exp, cmp.Diff(got, exp)))
		}
	}
}
