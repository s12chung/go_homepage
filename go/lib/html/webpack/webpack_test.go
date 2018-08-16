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

func defaultWebpack() (*Webpack, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewWebpack(generatedPath, log), hook
}

func TestAssetsPath(t *testing.T) {
	testCases := []struct {
		env          string
		fromPrevious bool
		expected     string
	}{
		{"", false, "assets"},
		{"", true, "assets"},
		{"test_env", true, "assets"},
		{"test_env", false, "test_env"},
		{"test_again", true, "test_env"},
		{"", true, "test_env"},
		{"test_again", false, "test_again"},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"env":          tc.env,
			"fromPrevious": tc.fromPrevious,
		})

		if !tc.fromPrevious {
			assetsPath = ""
		}
		os.Setenv("ASSETS_PATH", tc.env)

		got := AssetsPath()
		if got != tc.expected {
			t.Error(context.GotExpString("Result", got, tc.expected))
		}
	}
}

func TestAssetsUrl(t *testing.T) {
	assetsPath = "asset_test"
	got := AssetsUrl()
	test.AssertLabel(t, "Result", got, "/asset_test/")
}

func TestWebpack_GeneratedAssetsPath(t *testing.T) {
	assetsPath = "asset_test"
	webpack, _ := defaultWebpack()
	got := webpack.GeneratedAssetsPath()
	test.AssertLabel(t, "Result", got, path.Join(generatedPath, assetsPath))
}

func TestWebpack_GeneratedManifestPath(t *testing.T) {
	assetsPath = "asset_test"
	webpack, _ := defaultWebpack()
	got := webpack.GeneratedManifestPath()
	test.AssertLabel(t, "Result", got, path.Join(generatedPath, assetsPath, manifestpath))
}

func TestWebpack_ManifestUrl(t *testing.T) {
	assetsPath = "manifesttest"
	webpack, hook := defaultWebpack()
	got := webpack.ManifestUrl("vendor.css")

	test.PrintLogEntries(t, hook)
	test.AssertLabel(t, "Result", got, path.Join(assetsPath, "vendor-32267303b2484ed8b3aa.css"))
}

func TestWebpack_GetResponsiveImage(t *testing.T) {
	assetsPath = "responsivetest"
	webpack, hook := defaultWebpack()

	testCases := []struct {
		originalSrc string
		expected    *ResponsiveImage
	}{
		{"test.jpg",
			&ResponsiveImage{
				"responsivetest/content/images/test-37a65f446db3e9da33606b7eb48721bb-325.jpg 325w, responsivetest/content/images/test-c9d1dad468456287c20a476ade8a4d3f-750.jpg 750w, responsivetest/content/images/test-be268849aa760a62798817c27db7c430-1500.jpg 1500w, responsivetest/content/images/test-38e5ee006bf91e6af6d508bce2a9da4c-3000.jpg 3000w, responsivetest/content/images/test-84800b3286f76133d1592c9e68fa10be-4000.jpg 4000w",
				"responsivetest/content/images/test-37a65f446db3e9da33606b7eb48721bb-325.jpg",
			},
		},
		{"test.png",
			&ResponsiveImage{
				"responsivetest/content/images/test-afe607afeab81578d972f0ce9a92bdf4-325.png 325w, responsivetest/content/images/test-d31be3db558b4fe54b2c098abdd96306-750.png 750w, responsivetest/content/images/test-e4b7c37523ea30081ad02f6191b299f6-1440.png 1440w",
				"responsivetest/content/images/test-afe607afeab81578d972f0ce9a92bdf4-325.png",
			},
		},
	}

	for _, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"originalSrc": tc.originalSrc,
		})

		got := webpack.GetResponsiveImage(tc.originalSrc)

		test.PrintLogEntries(t, hook)
		if !cmp.Equal(got, tc.expected) {
			t.Error(context.GotExpString("result", got, tc.expected))
		}
	}
}
