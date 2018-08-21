package webpack

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/lib/utils"
	"github.com/s12chung/go_homepage/go/test"
)

func copyResponsiveImage(v *ResponsiveImage) *ResponsiveImage {
	cp := *v
	return &cp
}

func TestResponsiveImage_ChangeSrcPrefix(t *testing.T) {
	placeholder := "PREFIX/"

	testCases := []struct {
		img     *ResponsiveImage
		exp     *ResponsiveImage
		safeLog bool
	}{
		{
			&ResponsiveImage{"", ""},
			&ResponsiveImage{"", ""},
			true,
		},
		{
			&ResponsiveImage{"blah.png", ""},
			&ResponsiveImage{placeholder + "blah.png", ""},
			true,
		},
		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			true,
		},
		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w, blah-125.png 250w, blah-125.png 125w, blah-500.png 500w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, PREFIX/blah-125.png 250w, PREFIX/blah-125.png 125w, PREFIX/blah-500.png 500w"},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", ""},
			&ResponsiveImage{placeholder + "blah.png", ""},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w, content/images/blah-125.png 250w, content/images/blah-125.png 125w, content/images/blah-500.png 500w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, PREFIX/blah-125.png 250w, PREFIX/blah-125.png 125w, PREFIX/blah-500.png 500w"},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			false,
		}, {
			&ResponsiveImage{"content/images/blah.png", ",content/images/blah-125.png 125w,,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			false,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w, content/images/blah-125.png 250w, content/images/blah-125.png 125w, content/images/blah-500.png 500w,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, PREFIX/blah-125.png 250w, PREFIX/blah-125.png 125w, PREFIX/blah-500.png 500w"},
			false,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w, , content/images/blah-125.png 250w, , content/images/blah-125.png 125w, content/images/blah-500.png 500w,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, PREFIX/blah-125.png 250w, PREFIX/blah-125.png 125w, PREFIX/blah-500.png 500w"},
			false,
		},
	}

	for testCaseIndex, tc := range testCases {
		prefixes := []string{
			"",
			"/",
			"testy",
			"/testy",
			"testy/",
			"/testy/",
			"long/long",
			"long/long/way/",
		}

		for _, prefix := range prefixes {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":  testCaseIndex,
				"img":    tc.img,
				"prefix": prefix,
			})

			log, hook := logTest.NewNullLogger()
			got := copyResponsiveImage(tc.img)
			got.ChangeSrcPrefix(prefix, log)

			exp := copyResponsiveImage(tc.exp)
			cleanPrefix := utils.CleanFilePath(prefix)
			if cleanPrefix != "" {
				cleanPrefix = cleanPrefix + "/"
			}
			exp.Src = strings.Replace(exp.Src, placeholder, cleanPrefix, 1)
			exp.SrcSet = strings.Replace(exp.SrcSet, placeholder, cleanPrefix, -1)

			if !cmp.Equal(got, exp) {
				t.Error(context.GotExpString("Result", got, exp))
			}
			if test.SafeLogEntries(hook) != tc.safeLog {
				t.Error(context.GotExpString("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), tc.safeLog))
			}
		}
	}
}

func TestResponsiveImage_HtmlAttrs(t *testing.T) {
	testCases := []struct {
		img *ResponsiveImage
		exp string
	}{
		{
			&ResponsiveImage{"", ""},
			"",
		},
		{
			&ResponsiveImage{"blah.png", ""},
			`src="blah.png"`,
		},

		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w"},
			`src="blah.png" srcset="blah-125.png 125w"`,
		},
		{
			&ResponsiveImage{"", "blah-125.png 125w"},
			`srcset="blah-125.png 125w"`,
		},
		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w, blah-125.png 250w, blah-125.png 125w, blah-500.png 500w"},
			`src="blah.png" srcset="blah-125.png 125w, blah-125.png 250w, blah-125.png 125w, blah-500.png 500w"`,
		},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index": testCaseIndex,
			"img":   tc.img,
		})

		got := tc.img.HtmlAttrs()
		if got != tc.exp {
			t.Error(context.GotExpString("Result", got, tc.exp))
		}
	}
}
