package goodreads

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ernsheong/grand"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/go_homepage/go/test"
)

func defaultClient(t *testing.T, apiUrl string) (*Client, *logTest.Hook, func()) {
	log, hook := logTest.NewNullLogger()
	newCachePath, clean := test.SandboxDir(t, DefaultSettings().CachePath)
	settings := TestSettings(newCachePath, apiUrl)
	return NewClient(settings, log), hook, clean
}

func invalidateSettings(client *Client) {
	InvalidateSettings(client.Settings)
}

type serverSettings struct {
	failRequest bool
	context     *test.Context

	tracker *serverTracker
	server  *httptest.Server
}

func newServerSettings() *serverSettings {
	return &serverSettings{
		failRequest: false,
		context:     test.NewContext(),
		tracker:     newServerTracker(),
	}
}

type serverTracker struct {
	perPages         []int
	numberOfRequests int
}

func newServerTracker() *serverTracker {
	return &serverTracker{
		make([]int, 0),
		0,
	}
}

var defaultParamChecks = map[string]string{
	"id":    "1",
	"key":   "good_test",
	"shelf": "read",
	"sort":  "date_added",
	"order": "d",
}

func newServer(t *testing.T, paramChecks map[string]string) *serverSettings {
	if paramChecks == nil {
		paramChecks = make(map[string]string)
	}
	for k, v := range defaultParamChecks {
		_, contains := paramChecks[k]
		if contains {
			continue
		}
		paramChecks[k] = v
	}

	ss := newServerSettings()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		for key, exp := range paramChecks {
			if query[key][0] != exp {
				t.Error(ss.context.GotExpString("server.query."+key, query[key][0], exp))
			}
		}

		exp := "/review/list"
		if r.URL.Path != exp {
			t.Error(ss.context.GotExpString("r.URL.Path", r.URL.Path, exp))
		}

		perPage, err := strconv.ParseInt(query["per_page"][0], 10, 32)
		if err != nil {
			t.Error(ss.context.String(err))
		}
		ss.tracker.perPages = append(ss.tracker.perPages, int(perPage))
		ss.tracker.numberOfRequests += 1

		if ss.failRequest {
			return
		}

		page := query["page"][0]
		bytes := test.ReadFixture(t, fmt.Sprintf("page%v.xml", page))
		w.Write(bytes)
	}))
	ss.server = server
	return ss
}

func TestRatingMap(t *testing.T) {
	exp := map[int]int{
		1: 3,
		2: 5,
		3: 9,
		4: 0,
		5: 1,
	}

	var books []*Book

	for rating, count := range exp {
		for i := 0; i < count; i++ {
			books = append(books, &Book{Rating: rating})
		}
	}

	rand.Shuffle(len(books), func(i, j int) {
		books[i], books[j] = books[j], books[i]
	})

	got := RatingMap(books)
	if !cmp.Equal(got, exp) {
		t.Errorf("got: %v, exp: %v", got, exp)
	}
}

func TestClient_GetBooks(t *testing.T) {
	ss := newServer(t, nil)
	defer ss.server.Close()

	testCases := []struct {
		invalidSettings   bool
		failRequest       bool
		perPage           int
		numberOfRequests  int
		emptyBooks        bool
		unsafeLogEntries  bool
		fromPreviousCache bool
	}{
		{invalidSettings: true, emptyBooks: true, unsafeLogEntries: true},
		{failRequest: true, perPage: DefaultSettings().MaxPerPage, numberOfRequests: 1, emptyBooks: true, unsafeLogEntries: true},
		{perPage: DefaultSettings().MaxPerPage, numberOfRequests: 3},
		{perPage: DefaultSettings().MaxPerPage, numberOfRequests: 3},
		{perPage: DefaultSettings().PerPage, numberOfRequests: 1, fromPreviousCache: true},
		{perPage: DefaultSettings().PerPage, numberOfRequests: 1, fromPreviousCache: true},
		{invalidSettings: true, unsafeLogEntries: true, fromPreviousCache: true},
		{perPage: DefaultSettings().MaxPerPage, numberOfRequests: 3},
	}

	var client *Client
	var cleans []func()
	defer func() {
		for _, clean := range cleans {
			clean()
		}
	}()

	for testCaseIndex, tc := range testCases {
		ss.tracker = newServerTracker()
		ss.failRequest = tc.failRequest
		ss.context.SetFields(test.ContextFields{
			"index":             testCaseIndex,
			"invalidSettings":   tc.invalidSettings,
			"failRequest":       tc.failRequest,
			"fromPreviousCache": tc.fromPreviousCache,
		})

		newClient, hook, clean := defaultClient(t, ss.server.URL)
		cleans = append(cleans, clean)
		if tc.fromPreviousCache {
			newClient.Settings.CachePath = client.Settings.CachePath
		}
		client = newClient
		if tc.invalidSettings {
			invalidateSettings(client)
		}

		books, err := client.GetBooks()

		tracker := ss.tracker
		expPerPages := make([]int, tracker.numberOfRequests)
		for i := range expPerPages {
			expPerPages[i] = tc.perPage
		}
		if !cmp.Equal(tracker.perPages, expPerPages) {
			t.Error(ss.context.GotExpString("tracker.perPages", tracker.perPages, expPerPages))
		}

		if tracker.numberOfRequests != tc.numberOfRequests {
			t.Error(ss.context.GotExpString("tracker.numberOfRequests", tracker.numberOfRequests, tc.numberOfRequests))
		}

		if test.SafeLogEntries(hook) != !tc.unsafeLogEntries {
			t.Error(ss.context.GotExpString("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), !tc.unsafeLogEntries))
			test.PrintLogEntries(t, hook)
		}
		if !tc.unsafeLogEntries {
			cachePath := client.Settings.CachePath
			_, err := os.Stat(cachePath)
			if err != nil {
				t.Error(ss.context.String(err))
			}
			_, err = os.Stat(filepath.Join(cachePath, booksCacheFilename))
			if err != nil {
				t.Error(ss.context.String(err))
			}
		}

		if len(books) > 0 != !tc.emptyBooks {
			t.Error(ss.context.GotExpString("len(books) > 0", len(books) > 0, !tc.emptyBooks))
		}

		if err != nil {
			t.Error(ss.context.GotExpString("err", err, nil))
		}
	}
}

func TestClient_GetBooksRequest(t *testing.T) {
	testCases := []struct {
		userId               int
		numberOfInitialBooks int
		numberOfRequests     int
		unsafeLogEntries     bool
		failRequest          bool
	}{
		{5, 0, 3, false, false},
		{5, 20, 2, false, false},
		{5, 20, 1, true, true},
	}

	for testCaseIndex, tc := range testCases {
		ss := newServer(t, map[string]string{"id": strconv.Itoa(tc.userId)})
		ss.failRequest = tc.failRequest
		ss.context.SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"failRequest": tc.failRequest,
		})
		client, hook, clean := defaultClient(t, ss.server.URL)

		bookMap := make(map[string]*Book)
		test.RandSeed()
		for i := 0; i < tc.numberOfInitialBooks; i++ {
			id := grand.GenerateRandomString(10)
			bookMap[id] = &Book{Id: id}
		}
		client.GetBooksRequest(tc.userId, bookMap)

		tracker := ss.tracker
		if tracker.numberOfRequests != tc.numberOfRequests {
			t.Error(ss.context.GotExpString("tracker.numberOfRequests", tracker.numberOfRequests, tc.numberOfRequests))
		}
		if test.SafeLogEntries(hook) != !tc.unsafeLogEntries {
			t.Error(ss.context.GotExpString("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), !tc.unsafeLogEntries))
			test.PrintLogEntries(t, hook)
		}

		pdt := time.FixedZone("PDT", int((-7 * time.Hour).Seconds()))
		bookId := "2474898895"
		dateAdded := time.Date(2018, 7, 29, 15, 39, 52, 0, pdt)
		dateUpdated := time.Date(2018, 7, 29, 15, 39, 54, 0, pdt)
		expBook := &Book{
			xml.Name{},
			bookId,
			"The Organization Man: The Book That Defined a Generation",
			[]string{"William H. Whyte", "Carlota Pérez"},
			"0812218191",
			"9780812218190",
			2,
			GoodreadsDate(dateAdded),
			GoodreadsDate(dateUpdated),
			dateAdded,
			dateUpdated,
		}

		gotBook, contains := bookMap[bookId]
		if !contains {
			if !tc.failRequest {
				t.Error(ss.context.Stringf("bookMap[bookId] does not contain %v in %v", bookId, bookMap))
			}
		} else {
			if !cmp.Equal(gotBook, expBook, cmpopts.IgnoreFields(Book{}, "XMLName")) {
				t.Error(ss.context.GotExpString("bookMap[bookId]", gotBook, expBook))
			}
		}

		clean()
		ss.server.Close()
	}
}

func TestClient_GetBooksRequest_RateLimit(t *testing.T) {
	testCases := []struct {
		rateLimit             int
		lessThanQuarterSecond bool
	}{
		{1, true},
		{125, false},
	}

	for testCaseIndex, tc := range testCases {
		ss := newServer(t, nil)
		ss.context.SetFields(test.ContextFields{
			"index":     testCaseIndex,
			"rateLimit": tc.rateLimit,
		})

		client, _, clean := defaultClient(t, ss.server.URL)
		client.Settings.RateLimit = tc.rateLimit

		diff := test.TimeDiff(func() {
			client.GetBooksRequest(1, map[string]*Book{})
		})

		context := ss.context
		got := diff.Seconds() < 0.25
		if got != tc.lessThanQuarterSecond {
			t.Error(context.GotExpString("diff.Seconds() < 0.5", got, tc.lessThanQuarterSecond))
		}

		exp := 3
		if ss.tracker.numberOfRequests != exp {
			t.Error(context.GotExpString("ss.tracker.numberOfRequests", ss.tracker.numberOfRequests, exp))
		}

		ss.server.Close()
		clean()
	}
}

func TestBook_ReviewString(t *testing.T) {
	date := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

	book := &Book{
		xml.Name{},
		"1",
		"the book title",
		[]string{"berry", "jerry", "daisy"},
		"zzzzz",
		"yyyyy",
		4,
		GoodreadsDate(date),
		GoodreadsDate(date),
		date,
		date,
	}

	got := book.ReviewString()
	exp := `"the book title" by berry, jerry & daisy ****`

	if got != exp {
		t.Errorf("got: %v, exp: %v", got, exp)
	}
}
