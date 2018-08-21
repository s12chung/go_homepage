package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/go_homepage/go/test"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/router"
	"github.com/s12chung/go_homepage/go/test/mocks"
	"path"
)

//go:generate mockgen -destination=../../test/mocks/routes_helper.go -package=mocks github.com/s12chung/go_homepage/go/content/routes Helper

func modelsConfig() {
	log, _ := logTest.NewNullLogger()
	models.TestConfig(path.Join("../models", test.FixturePath), log)
}

func setPostDirEmpty() {
	log, _ := logTest.NewNullLogger()
	models.TestSetPostDirEmpty(log)
}

func goodreadsSettings(t *testing.T, apiUrl string) (*goodreads.Settings, func()) {
	newCachePath, clean := test.SandboxDir(t, goodreads.DefaultSettings().CachePath)
	settings := goodreads.TestSettings(newCachePath, apiUrl)
	return settings, clean
}

func TestMain(m *testing.M) {
	modelsConfig()
	retCode := m.Run()
	os.Exit(retCode)
}

func defaultAllRoutes() *AllRoutes {
	return NewAllRoutes(NewBaseHelper(nil, nil, nil, nil))
}

func TestAllRoutes_WildcardUrls(t *testing.T) {
	testCases := []struct {
		postDirEmpty bool
		expected     []string
	}{
		{true, []string{}},
		{false, []string{"/post1", "/post2", "/draft1", "/draft2", "/draft3"}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"postDirEmpty": tc.postDirEmpty,
		})

		modelsConfig()
		if tc.postDirEmpty {
			setPostDirEmpty()
		}

		allRoutes := defaultAllRoutes()
		got, err := allRoutes.WildcardUrls()
		if err != nil {
			t.Error(context.String(err))
		}
		if !cmp.Equal(got, tc.expected) {
			t.Error(context.GotExpString("Result", got, tc.expected))
		}
	}
}

func testRoute(t *testing.T, callback func(helper *mocks.MockHelper, ctx *mocks.MockContext)) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	ctx := mocks.NewMockContext(controller)
	helper := mocks.NewMockHelper(controller)
	callback(helper, ctx)
}

func TestAllRoutes_getAbout(t *testing.T) {
	testRoute(t, func(helper *mocks.MockHelper, ctx *mocks.MockContext) {
		helper.EXPECT().RespondUrlHTML(ctx, nil)
		err := NewAllRoutes(helper).getAbout(ctx)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestAllRoutes_getReading(t *testing.T) {
	testCases := []struct {
		emptyResponse bool
		years         []int
		ratingMap     map[int]int
	}{
		{true, []int{}, map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}},
		{false, []int{2000, 2010, 2018}, map[int]int{1: 0, 2: 1, 3: 0, 4: 2, 5: 0}},
	}

	for testCaseIndex, tc := range testCases {
		testRoute(t, func(helper *mocks.MockHelper, ctx *mocks.MockContext) {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":         testCaseIndex,
				"emptyResponse": tc.emptyResponse,
			})

			log, _ := logTest.NewNullLogger()
			ctx.EXPECT().Log().Return(log)

			emptyResponse := tc.emptyResponse
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if emptyResponse {
					return
				}
				w.Write(test.ReadFixture(t, "goodreads.xml"))
			}))
			defer server.Close()
			settings, clean := goodreadsSettings(t, server.URL)
			defer clean()

			helper.EXPECT().GoodreadsSettings().Return(settings)
			helper.EXPECT().RespondUrlHTML(ctx, gomock.Any()).Do(func(ctx router.Context, data interface{}) {
				d, ok := data.(readingData)
				if !ok {
					t.Error(context.Stringf("could not convert to: %v", readingData{}))
				}
				years := make([]int, len(d.Books))
				for i, book := range d.Books {
					years[i] = book.SortedDate().Year()
				}
				sort.Ints(years)
				sort.Ints(tc.years)
				if !cmp.Equal(years, tc.years) {
					t.Error(context.GotExpString("years", years, tc.years))
				}

				d.Books = nil
				earliestYear := time.Now().Year()
				if len(tc.years) != 0 {
					earliestYear = tc.years[0]
				}
				exp := readingData{
					nil,
					tc.ratingMap,
					earliestYear,
				}
				if !cmp.Equal(d, exp) {
					t.Error(context.GotExpString("d", d, exp))
				}
			})

			err := NewAllRoutes(helper).getReading(ctx)
			if err != nil {
				t.Error(context.String(err))
			}
		})
	}
}

func TestAllRoutes_getPost(t *testing.T) {
	testCases := []struct {
		postFilename string
		exists       bool
	}{
		{"draft1", true},
		{"post1", true},
		{"does not exist", false},
	}

	for testCaseIndex, tc := range testCases {
		testRoute(t, func(helper *mocks.MockHelper, ctx *mocks.MockContext) {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":        testCaseIndex,
				"postFilename": tc.postFilename,
			})

			ctx.EXPECT().UrlParts().Return([]string{tc.postFilename})
			if tc.exists {
				helper.EXPECT().RespondHTML(ctx, "post", gomock.Any()).Do(func(ctx router.Context, templateName string, data interface{}) {
					d, ok := data.(*models.Post)
					if !ok {
						t.Error(context.Stringf("could not convert to: %v", models.Post{}))
					}
					if d.Id() != tc.postFilename {
						t.Error(context.GotExpString("Wrong Post", d.Id(), tc.postFilename))
					}
				})
			}

			err := NewAllRoutes(helper).getPost(ctx)
			if !tc.exists {
				if err == nil {
					t.Error(context.String("no error for not existing"))
				}
				return
			}
			if err != nil {
				t.Error(context.String(err))
			}
		})
	}
}

func TestAllRoutes_getPosts(t *testing.T) {
	testCases := []struct {
		postDirEmpty bool
		expected     []string
	}{
		{true, []string{}},
		{false, []string{"post1", "post2"}},
	}

	for testCaseIndex, tc := range testCases {
		testRoute(t, func(helper *mocks.MockHelper, ctx *mocks.MockContext) {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":        testCaseIndex,
				"postDirEmpty": tc.postDirEmpty,
			})

			modelsConfig()
			if tc.postDirEmpty {
				setPostDirEmpty()
			}

			helper.EXPECT().RespondHTML(ctx, "posts", gomock.Any()).Do(func(ctx router.Context, templateName string, data interface{}) {
				d, ok := data.(postsData)
				if !ok {
					t.Error(context.Stringf("could not convert to: %v", postsData{}))
				}
				ids := make([]string, len(d.Posts))
				for i, post := range d.Posts {
					ids[i] = post.Id()
				}
				sort.Strings(ids)
				sort.Strings(tc.expected)
				if !cmp.Equal(ids, tc.expected) {
					t.Error(context.GotExpString("ids", ids, tc.expected))
				}
			})

			err := NewAllRoutes(helper).getPosts(ctx)
			if err != nil {
				t.Error(context.String(err))
			}
		})
	}
}

func TestAllRoutes_getPostsAtom(t *testing.T) {
	testCases := []struct {
		postDirEmpty bool
		expected     []string
	}{
		{true, []string{}},
		{false, []string{"post1", "post2"}},
	}

	for testCaseIndex, tc := range testCases {
		testRoute(t, func(helper *mocks.MockHelper, ctx *mocks.MockContext) {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":        testCaseIndex,
				"postDirEmpty": tc.postDirEmpty,
			})

			modelsConfig()
			if tc.postDirEmpty {
				setPostDirEmpty()
			}

			expLogoUrl := "test_logo.png"
			helper.EXPECT().ManifestUrl("images/logo.png").Return(expLogoUrl)
			helper.EXPECT().RespondAtom(ctx, "posts", expLogoUrl, gomock.Any()).
				Do(func(tx router.Context, feedName, logoUrl string, htmlEntries []*atom.HtmlEntry) {

					ids := make([]string, len(htmlEntries))
					for i, htmlEntry := range htmlEntries {
						ids[i] = htmlEntry.Id
					}
					sort.Strings(ids)
					sort.Strings(tc.expected)
					if !cmp.Equal(ids, tc.expected) {
						t.Error(context.GotExpString("ids", ids, tc.expected))
					}
				})

			err := NewAllRoutes(helper).getPostsAtom(ctx)
			if err != nil {
				t.Error(context.String(err))
			}
		})
	}
}

func TestAllRoutes_getRobotsTxt(t *testing.T) {
	testRoute(t, func(helper *mocks.MockHelper, ctx *mocks.MockContext) {
		ctx.EXPECT().Respond([]byte{})
		err := NewAllRoutes(helper).getRobotsTxt(ctx)
		if err != nil {
			t.Error(err)
		}
	})
}
