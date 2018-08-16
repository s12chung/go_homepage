package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
	"path"
	"sort"
)

func TestCleanFilePath(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"/go/src/github.com/s12chung/go_homepage", "go/src/github.com/s12chung/go_homepage"},
		{"/go/src/github.com/s12chung/go_homepage/", "go/src/github.com/s12chung/go_homepage"},
		{"go/src/github.com/s12chung/go_homepage", "go/src/github.com/s12chung/go_homepage"},
		{"./go/src/github.com/s12chung/go_homepage", "go/src/github.com/s12chung/go_homepage"},
		{"./../go/src/github.com/s12chung/go_homepage", "../go/src/github.com/s12chung/go_homepage"},
		{"", ""},
		{"./", ""},
		{".", ""},
	}

	for _, tc := range testCases {
		got := CleanFilePath(tc.input)
		test.AssertInput(t, tc.input, got, tc.expected)
	}
}

func TestToSimpleQuery(t *testing.T) {
	testCases := []struct {
		input    map[string]string
		expected string
	}{
		{map[string]string{"a": "1", "b": "2", "c": "3"}, "a=1&b=2&c=3"},
		{map[string]string{"a": "1"}, "a=1"},
		{map[string]string{}, ""},
	}

	for _, tc := range testCases {
		got := ToSimpleQuery(tc.input)
		test.AssertInput(t, tc.input, got, tc.expected)
	}
}

func TestSliceList(t *testing.T) {
	testCases := []struct {
		input    []string
		expected string
	}{
		{[]string{"Johnny", "Eugene", "Kate", "Katherine"}, "Johnny, Eugene, Kate & Katherine"},
		{[]string{"Mike", "Cedric"}, "Mike & Cedric"},
		{[]string{"Steve"}, "Steve"},
		{[]string{}, ""},
	}

	for _, tc := range testCases {
		got := SliceList(tc.input)
		test.AssertInput(t, tc.input, got, tc.expected)
	}
}

func TestFilePaths(t *testing.T) {
	testCases := []struct {
		suffix   string
		dirPaths []string
		expected map[string][]string
		error    bool
	}{
		{".md", []string{""}, map[string][]string{"": {"a.md", "b.md"}}, false},
		{".md", []string{"dir1"}, map[string][]string{"dir1": {"1.md"}}, false},
		{".md", []string{"dir1", "dir2"}, map[string][]string{"dir1": {"1.md"}}, false},
		{".md", []string{"dir1", "dir2", "dir3"}, map[string][]string{"dir1": {"1.md"}}, false},
		{".md", []string{"", "dir1", "dir2", "dir3"}, map[string][]string{"": {"a.md", "b.md"}, "dir1": {"1.md"}}, false},
		{".txt", []string{""}, map[string][]string{}, false},
		{".txt", []string{"dir1"}, map[string][]string{"dir1": {"1.txt", "2.txt"}}, false},
		{".txt", []string{"dir1", "dir2"}, map[string][]string{"dir1": {"1.txt", "2.txt"}, "dir2": {"a.txt"}}, false},
		{".txt", []string{"dir1", "dir2", "dir3"}, map[string][]string{"dir1": {"1.txt", "2.txt"}, "dir2": {"a.txt"}}, false},
		{".txt", []string{"", "dir1", "dir2", "dir3"}, map[string][]string{"dir1": {"1.txt", "2.txt"}, "dir2": {"a.txt"}}, false},
		{".txt", []string{"does not exist"}, nil, true},
		{".md", []string{"", "does not exist"}, nil, true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":    testCaseIndex,
			"suffix":   tc.suffix,
			"dirPaths": tc.dirPaths,
		})

		dirPaths := make([]string, len(tc.dirPaths))
		for i, d := range tc.dirPaths {
			dirPaths[i] = path.Join(test.FixturePath, d)
		}

		got, err := FilePaths(tc.suffix, dirPaths...)
		if tc.error && err != nil {
			continue
		}
		if err != nil {
			t.Error(context.String(err))
			continue
		}

		var exp []string
		for relativePath, files := range tc.expected {
			for _, file := range files {
				exp = append(exp, path.Join(test.FixturePath, relativePath, file))
			}
		}

		sort.Strings(exp)
		sort.Strings(got)

		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}

func TestGetStringField(t *testing.T) {
	type testCase struct {
		data     interface{}
		dataPtr  interface{}
		name     string
		expected string
	}
	newTestCase := func(name, expected string, dataF func() (interface{}, interface{})) testCase {
		data, dataPtr := dataF()
		return testCase{data, dataPtr, name, expected}
	}

	testCases := []testCase{
		newTestCase("test", "", func() (interface{}, interface{}) {
			return nil, nil
		}),
		newTestCase("Name", "", func() (interface{}, interface{}) {
			data := struct{}{}
			return data, &data
		}),
		newTestCase("Name", "the name", func() (interface{}, interface{}) {
			data := struct{ Name string }{"the name"}
			return data, &data
		}),
		newTestCase("Name", "the name", func() (interface{}, interface{}) {
			data := struct {
				Name string
				Zors string
			}{"the name", "zor"}
			return data, &data
		}),
		newTestCase("Name", "the name", func() (interface{}, interface{}) {
			data := struct {
				Name  string
				Which []int
			}{"the name", []int{1, 2}}
			return data, &data
		}),
		newTestCase("Which", "", func() (interface{}, interface{}) {
			data := struct {
				Name  string
				Which []int
			}{"the name", []int{1, 2}}
			return data, &data
		}),
		newTestCase("Name", "", func() (interface{}, interface{}) {
			data := struct {
				Name  []int
				Zors  string
				Which []int
			}{[]int{100, 200, 300}, "zor", []int{1, 2}}
			return data, &data
		}),
		newTestCase("Zors", "zor", func() (interface{}, interface{}) {
			data := struct {
				Name  []int
				Zors  string
				Which []int
			}{[]int{100, 200, 300}, "zor", []int{1, 2}}
			return data, &data
		}),
		newTestCase("NoExist", "", func() (interface{}, interface{}) {
			data := struct {
				Name  []int
				Zors  string
				Which []int
			}{[]int{100, 200, 300}, "zor", []int{1, 2}}
			return data, &data
		}),
		newTestCase("No", "", func() (interface{}, interface{}) {
			data := struct {
				Name  []int
				Zors  string
				Which []int
			}{[]int{100, 200, 300}, "zor", []int{1, 2}}
			return data, &data
		}),
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index": testCaseIndex,
			"data":  tc.data,
			"name":  tc.name,
		})

		got := GetStringField(tc.data, tc.name)
		if got != tc.expected {
			t.Error(context.GotExpString("Struct", got, tc.expected))
		}
		got = GetStringField(tc.dataPtr, tc.name)
		if got != tc.expected {
			t.Error(context.GotExpString("Pointer", got, tc.expected))
		}
	}
}
