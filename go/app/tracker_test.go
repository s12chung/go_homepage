package app

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
		error           bool
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
		{[]string{"a", "b"}, true, []string{"c"}, []string{}},
		{[]string{"a", "b"}, true, []string{"b", "c"}, []string{}},
		{[]string{"a", "b"}, true, []string{"a", "b", "c"}, []string{}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":         testCaseIndex,
			"allUrls":       tc.allUrls,
			"error":         tc.error,
			"dependentUrls": tc.dependentUrls,
		})

		tracker := defaultTracker(func() ([]string, error) {
			if tc.error {
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
			if tc.error == false {
				t.Error(context.String(err))
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
