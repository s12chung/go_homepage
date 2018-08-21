package webpack

import (
	"fmt"
	"os"
	"path"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
)

var generatedPath = path.Join(test.FixturePath, "generated")

var jpgResponsiveImage = &ResponsiveImage{
	"assets/test-37a65f446db3e9da33606b7eb48721bb-325.jpg",
	"assets/test-37a65f446db3e9da33606b7eb48721bb-325.jpg 325w, assets/test-c9d1dad468456287c20a476ade8a4d3f-750.jpg 750w, assets/test-be268849aa760a62798817c27db7c430-1500.jpg 1500w, assets/test-38e5ee006bf91e6af6d508bce2a9da4c-3000.jpg 3000w, assets/test-84800b3286f76133d1592c9e68fa10be-4000.jpg 4000w",
}
var pngResponsiveImage = &ResponsiveImage{
	"assets/test-afe607afeab81578d972f0ce9a92bdf4-325.png",
	"assets/test-afe607afeab81578d972f0ce9a92bdf4-325.png 325w, assets/test-d31be3db558b4fe54b2c098abdd96306-750.png 750w, assets/test-e4b7c37523ea30081ad02f6191b299f6-1440.png 1440w",
}

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
	webpack, hook := defaultWebpack()
	got := webpack.ManifestUrl("vendor.css")

	test.PrintLogEntries(t, hook)
	test.AssertLabel(t, "Result", got, path.Join(webpack.AssetsPath(), "vendor-32267303b2484ed8b3aa.css"))
}

func TestWebpack_GetResponsiveImage(t *testing.T) {
	webpack, hook := defaultWebpack()

	testCases := []struct {
		originalSrc      string
		expected         *ResponsiveImage
		badKey           bool
		notHasResponsive bool
	}{
		{"test.jpg", jpgResponsiveImage, false, false},
		{"test.png", pngResponsiveImage, false, false},
		{"test.png", &ResponsiveImage{Src: "test.png"}, true, false},
		{"test.gif", &ResponsiveImage{Src: "assets/test.gif"}, false, true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":            testCaseIndex,
			"originalSrc":      tc.originalSrc,
			"badKey":           tc.badKey,
			"notHasResponsive": tc.notHasResponsive,
		})

		key := "test"
		if tc.badKey {
			key = ""
		}
		got := webpack.GetResponsiveImage(key, tc.originalSrc)
		if !cmp.Equal(got, tc.expected) {
			t.Error(context.GotExpString("result", got, tc.expected))
		}
		if tc.badKey && test.SafeLogEntries(hook) {
			test.PrintLogEntries(t, hook)
			t.Error(context.String("expecting unsafe log entry"))
		}
	}
}

func TestWebpack_ReplaceResponsiveAttrs(t *testing.T) {
	complexImg := `<img alt="blah" src="test.jpg" class="haha"/>`
	testCases := []struct {
		responsiveKey string
		input         string
		expected      string
	}{
		{"", complexImg, complexImg},
		{"test", `<img src="test.jpg"/>`, fmt.Sprintf(`<img %v/>`, jpgResponsiveImage.HtmlAttrs())},
		{"test", `<img src="test.jpg" class="haha"/>`, fmt.Sprintf(`<img %v class="haha"/>`, jpgResponsiveImage.HtmlAttrs())},
		{"test", `<img alt="blah" src="test.jpg"/>`, fmt.Sprintf(`<img alt="blah" %v/>`, jpgResponsiveImage.HtmlAttrs())},
		{"test", complexImg, fmt.Sprintf(`<img alt="blah" %v class="haha"/>`, jpgResponsiveImage.HtmlAttrs())},
	}
	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":         testCaseIndex,
			"responsiveKey": tc.responsiveKey,
			"input":         tc.input,
		})

		webpack, hook := defaultWebpack()
		webpack.settings.ReplaceResponsiveAttrs = tc.responsiveKey

		got := webpack.ReplaceResponsiveAttrs(tc.input)
		if got != tc.expected {
			t.Error(context.GotExpString("result", got, tc.expected))
		}
		if tc.responsiveKey == "" && test.SafeLogEntries(hook) {
			test.PrintLogEntries(t, hook)
			t.Error(context.String("expecting unsafe log entry"))
		}
	}
}

func TestWebpack_ResponsiveHtmlAttrs(t *testing.T) {
	webpack, _ := defaultWebpack()
	got := string(webpack.ResponsiveHtmlAttrs("test", "test.jpg"))
	exp := jpgResponsiveImage.HtmlAttrs()
	if got != exp {
		t.Error(test.AssertLabelString("result", got, exp))
	}
}
