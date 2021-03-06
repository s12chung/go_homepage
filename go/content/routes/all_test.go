package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/test"

	"github.com/s12chung/gostatic-packages/atom"
	"github.com/s12chung/gostatic-packages/goodreads"

	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/test/mocks"
)

//go:generate mockgen -destination=../../test/mocks/router_context.go -package=mocks github.com/s12chung/gostatic/go/lib/router Context
//go:generate mockgen -destination=../../test/mocks/routes_helper.go -package=mocks github.com/s12chung/go_homepage/go/content/routes Helper

func modelsConfig() {
	log, _ := logTest.NewNullLogger()
	models.TestConfig(path.Join("../models", test.FixturePath), log)
}

func setPostDirEmpty() {
	log, _ := logTest.NewNullLogger()
	models.TestSetPostDirEmpty(log)
}

func testGoodreadSettings(cachePath, apiURL string) *goodreads.Settings {
	settings := goodreads.DefaultSettings()
	settings.APIURL = apiURL
	settings.APIKey = "good_test"
	settings.UserID = 1
	settings.RateLimit = 1
	settings.CachePath = cachePath
	return settings
}

func goodreadsSettings(t *testing.T, apiURL string) (*goodreads.Settings, func()) {
	newCachePath, clean := test.SandboxDir(t, goodreads.DefaultSettings().CachePath)
	settings := testGoodreadSettings(newCachePath, apiURL)
	return settings, clean
}

func TestMain(m *testing.M) {
	modelsConfig()
	retCode := m.Run()
	os.Exit(retCode)
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
		ctx.EXPECT().URL().Return("/about")
		helper.EXPECT().RespondHTML(ctx, "/about", layoutData{"About", nil})

		err := NewAllRoutes(helper).getAbout(ctx)
		if err != nil {
			t.Error(err)
		}
	})
}

type readingTestCase struct {
	emptyResponse bool
	years         []int
	ratingMap     map[int]int
}

func TestAllRoutes_getReading(t *testing.T) {
	testCases := []readingTestCase{
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
				_, err := w.Write(test.ReadFixture(t, "goodreads.xml"))
				if err != nil {
					t.Error(context.String(err))
				}
			}))
			defer server.Close()
			settings, clean := goodreadsSettings(t, server.URL)
			defer clean()

			ctx.EXPECT().URL().Return("/reading")
			helper.EXPECT().GoodreadsSettings().Return(settings)
			helper.EXPECT().RespondHTML(ctx, "/reading", gomock.Any()).Do(testReadingResponseF(t, context, tc))

			err := NewAllRoutes(helper).getReading(ctx)
			if err != nil {
				t.Error(context.String(err))
			}
		})
	}
}

func testReadingResponseF(t *testing.T, context *test.Context, tc readingTestCase) func(ctx router.Context, templateName string, data interface{}) {
	return func(ctx router.Context, templateName string, data interface{}) {
		layoutD, ok := data.(layoutData)
		if !ok {
			t.Error(context.Stringf("could not convert to: %v", layoutData{}))
			return
		}
		if layoutD.Title != "Reading" {
			t.Error(context.GotExpString("layoutD.Title", layoutD.Title, "Reading"))
		}

		d, ok := layoutD.ContentData.(readingData)
		if !ok {
			t.Error(context.Stringf("could not convert to: %v", readingData{}))
			return
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

			if tc.exists {
				helper.EXPECT().RespondHTML(ctx, "post", gomock.Any()).Do(func(ctx router.Context, templateName string, data interface{}) {
					layoutD, ok := data.(layoutData)
					if !ok {
						t.Error(context.Stringf("could not convert to: %v", layoutData{}))
						return
					}
					post, ok := layoutD.ContentData.(*models.Post)
					if !ok {
						t.Error(context.Stringf("could not convert to: %v", models.Post{}))
						return
					}
					if layoutD.Title != post.Title {
						t.Error(context.GotExpString("layoutD.Title", layoutD.Title, post.Title))
					}

					if post.ID() != tc.postFilename {
						t.Error(context.GotExpString("Wrong Post", post.ID(), tc.postFilename))
					}
				})
			}

			err := NewAllRoutes(helper).getPostF(tc.postFilename)(ctx)
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
				layoutD, ok := data.(layoutData)
				if !ok {
					t.Error(context.Stringf("could not convert to: %v", layoutData{}))
					return
				}
				if layoutD.Title != "" {
					t.Error(context.GotExpString("layoutD.Title", layoutD.Title, ""))
				}

				d, ok := layoutD.ContentData.(postsData)
				if !ok {
					t.Error(context.Stringf("could not convert to: %v", postsData{}))
				}
				ids := make([]string, len(d.Posts))
				for i, post := range d.Posts {
					ids[i] = post.ID()
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

			expLogoURL := "test_logo.png"
			helper.EXPECT().ManifestURL("images/logo.png").Return(expLogoURL)
			helper.EXPECT().RespondAtom(ctx, "posts", expLogoURL, gomock.Any()).
				Do(func(tx router.Context, feedName, logoURL string, htmlEntries []*atom.HTMLEntry) {

					ids := make([]string, len(htmlEntries))
					for i, htmlEntry := range htmlEntries {
						ids[i] = htmlEntry.ID
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
