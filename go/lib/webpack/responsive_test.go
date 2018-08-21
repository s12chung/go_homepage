package webpack

import (
	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func defaultResponsive() (*Responsive, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewResponsive(generatedPath, DefaultSettings().AssetsPath, "", log), hook
}

func TestHasResponsive(t *testing.T) {
	testCases := []struct {
		originalSrc string
		exp         bool
	}{
		{"test.jpg", true},
		{"test.png", true},
		{"test.gif", false},
		{"test.svg", false},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
		})
		got := HasResponsive(tc.originalSrc)
		if got != tc.exp {
			t.Error(context.GotExpString("result", got, tc.exp))
		}
	}
}

func TestResponsive_GetResponsiveImage(t *testing.T) {
	testCases := []struct {
		originalSrc string
		imagePath   string
		exp         *ResponsiveImage
	}{
		{"test.jpg", "", jpgResponsiveImage},
		{"test.png", "", pngResponsiveImage},
		{"test.gif", "", &ResponsiveImage{Src: "test.gif"}},
		{"test.png", "does_not_exist", &ResponsiveImage{Src: "test.png"}},
		{"http://testy.com/test.png", "", &ResponsiveImage{Src: "http://testy.com/test.png"}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
			"imagePath":   tc.imagePath,
		})

		responsive, hook := defaultResponsive()
		responsive.imagePath = tc.imagePath
		got := responsive.GetResponsiveImage(tc.originalSrc)

		if tc.imagePath != "" && test.SafeLogEntries(hook) {
			t.Error(context.String("expecting unsafe log entry"))
		}
		if !cmp.Equal(got, tc.exp) {
			t.Error(context.GotExpString("result", got, tc.exp))
		}
	}
}
