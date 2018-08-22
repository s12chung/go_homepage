package app

import (
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
	"path"
)

func TestDefaultSettings(t *testing.T) {
	test.TestEnvSetting(t, "GENERATED_PATH", "./generated", func() string {
		return DefaultSettings().GeneratedPath
	})
}

type ContentSettings struct {
	NumberOfPosts int `json:"number_of_posts,omitempty"`
}

func TestSettingsFromFile(t *testing.T) {
	testCases := []struct {
		file    string
		content interface{}
		success bool
		safeLog bool
	}{
		{"settings.json", nil, true, true},
		{"settings.json", &ContentSettings{}, true, true},
		{"does_not_exist.json", &ContentSettings{}, false, false},
		{"broken_settings.json", &ContentSettings{}, false, false},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":   testCaseIndex,
			"file":    tc.file,
			"content": tc.content,
		})

		log, hook := logTest.NewNullLogger()
		settings := SettingsFromFile(path.Join(test.FixturePath, tc.file), tc.content, log)

		if tc.success {
			if tc.content != nil {
				contentExp := &ContentSettings{98}
				if !cmp.Equal(tc.content, contentExp) {
					t.Error(context.GotExpString("settings.Content", settings.Content, contentExp))
				}

				exp := DefaultSettings()
				exp.Concurrency = 5
				exp.GeneratedPath = "some_path"
				exp.Content = contentExp
				if !cmp.Equal(settings, exp) {
					t.Error(context.GotExpString("settings", settings, exp))
				}
			}

		}
		if test.SafeLogEntries(hook) != tc.safeLog {
			t.Error(context.GotExpString("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), tc.safeLog))
		}
	}

}
