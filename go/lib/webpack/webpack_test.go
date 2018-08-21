package webpack

import (
	"os"
	"path"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
)

var generatedPath = path.Join(test.FixturePath, "generated")

func setAssetsPath(val string, callback func()) {
	os.Setenv("ASSETS_PATH", val)
	callback()
	os.Setenv("ASSETS_PATH", "")
}

func defaultWebpack() (*Webpack, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	settings := DefaultSettings()
	settings.ResponsiveImageMap = map[string]string{
		"test": "",
	}
	return NewWebpack(generatedPath, settings, log), hook
}

func TestWebpack_AssetsPath(t *testing.T) {
	testCases := []struct {
		env      string
		expected string
	}{
		{"", "assets"},
		{"test_env", "test_env"},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index": testCaseIndex,
			"env":   tc.env,
		})

		setAssetsPath(tc.env, func() {
			os.Setenv("ASSETS_PATH", tc.env)
			webpack, _ := defaultWebpack()
			got := webpack.AssetsPath()
			if got != tc.expected {
				t.Error(context.GotExpString("Result", got, tc.expected))
			}
		})
	}
}

func TestWebpack_AssetsUrl(t *testing.T) {
	webpack, _ := defaultWebpack()
	got := webpack.AssetsUrl()
	test.AssertLabel(t, "Result", got, "/assets/")
}

func TestWebpack_GeneratedAssetsPath(t *testing.T) {
	webpack, _ := defaultWebpack()
	got := webpack.GeneratedAssetsPath()
	test.AssertLabel(t, "Result", got, path.Join(generatedPath, webpack.AssetsPath()))
}

func TestWebpack_ManifestUrl(t *testing.T) {
	setAssetsPath("manifesttest", func() {
		webpack, hook := defaultWebpack()
		got := webpack.ManifestUrl("vendor.css")

		test.PrintLogEntries(t, hook)
		test.AssertLabel(t, "Result", got, path.Join(webpack.AssetsPath(), "vendor-32267303b2484ed8b3aa.css"))
	})
}

func TestWebpack_GetResponsiveImage(t *testing.T) {
	var webpack *Webpack
	var hook *logTest.Hook
	setAssetsPath("responsivetest", func() {
		webpack, hook = defaultWebpack()
	})

	testCases := []struct {
		originalSrc string
		expected    *ResponsiveImage
	}{
		{"test.jpg",
			&ResponsiveImage{
				"responsivetest/test-37a65f446db3e9da33606b7eb48721bb-325.jpg 325w, responsivetest/test-c9d1dad468456287c20a476ade8a4d3f-750.jpg 750w, responsivetest/test-be268849aa760a62798817c27db7c430-1500.jpg 1500w, responsivetest/test-38e5ee006bf91e6af6d508bce2a9da4c-3000.jpg 3000w, responsivetest/test-84800b3286f76133d1592c9e68fa10be-4000.jpg 4000w",
				"responsivetest/test-37a65f446db3e9da33606b7eb48721bb-325.jpg",
			},
		},
		{"test.png",
			&ResponsiveImage{
				"responsivetest/test-afe607afeab81578d972f0ce9a92bdf4-325.png 325w, responsivetest/test-d31be3db558b4fe54b2c098abdd96306-750.png 750w, responsivetest/test-e4b7c37523ea30081ad02f6191b299f6-1440.png 1440w",
				"responsivetest/test-afe607afeab81578d972f0ce9a92bdf4-325.png",
			},
		},
	}

	for _, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"originalSrc": tc.originalSrc,
		})

		got := webpack.GetResponsiveImage("test", tc.originalSrc)
		test.PrintLogEntries(t, hook)
		if !cmp.Equal(got, tc.expected) {
			t.Error(context.GotExpString("result", got, tc.expected))
		}
	}
}
