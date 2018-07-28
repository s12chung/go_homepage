package goodreads

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/utils"
)

const booksFile = "books.json"

type Client struct {
	Settings *settings.GoodreadsSettings
	log      logrus.FieldLogger
}

func NewClient(settings *settings.GoodreadsSettings, log logrus.FieldLogger) *Client {
	return &Client{
		settings,
		log,
	}
}

func (client *Client) GetBooks() ([]*Book, error) {
	if client.invalidSettings() {
		client.log.Warn("Invalid goodsreads Settings, skipping goodsreads API calls")
		return nil, nil
	}
	err := client.setup()
	if err != nil {
		return nil, err
	}

	bookMap, err := client.getBooks(client.Settings.UserId)
	if err != nil {
		return nil, err
	}
	return toBooks(bookMap), nil
}

func toBooks(bookMap map[string]*Book) []*Book {
	books := make([]*Book, len(bookMap))
	i := 0
	for _, book := range bookMap {
		books[i] = book
		i += 1
	}
	return books
}

func (client *Client) invalidSettings() bool {
	return client.Settings.ApiKey == "" && client.Settings.UserId == 0
}

func (client *Client) setup() error {
	cachePath := client.Settings.CachePath
	return os.MkdirAll(cachePath, 0755)
}

type GoodreadsDate time.Time

func (date *GoodreadsDate) UnmarshalXML(decoder *xml.Decoder, startElement xml.StartElement) error {
	var stringValue string

	err := decoder.DecodeElement(&stringValue, &startElement)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RubyDate, stringValue)
	if err != nil {
		return err
	}

	*date = GoodreadsDate(t)
	return nil
}

type Book struct {
	XMLName xml.Name `xml:"review" json:"-"`
	Id      string   `xml:"id" json:"id"`
	Title   string   `xml:"book>title" json:"title"`
	Authors []string `xml:"book>authors>author>name" json:"authors"`
	Isbn    string   `xml:"book>isbn" json:"isbn"`
	Isbn13  string   `xml:"book>isbn13" json:"isbn13"`
	Rating  int      `xml:"rating" json:"rating"`

	XMLDateAdded    GoodreadsDate `xml:"date_added" json:"-"`
	XXMLDateUpdated GoodreadsDate `xml:"date_updated" json:"-"`
	DateAdded       time.Time     `xml:"-" json:"date_added"`
	DateUpdated     time.Time     `xml:"-" json:"date_updated"`
}

func RatingMap(books []*Book) map[int]int {
	ratingMap := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	i := 0
	for _, book := range books {
		ratingMap[book.Rating] += 1
		i += 1
	}
	return ratingMap
}

func (book *Book) convertDates() {
	book.DateAdded = time.Time(book.XMLDateAdded)
	book.DateUpdated = time.Time(book.XXMLDateUpdated)
}

func (book *Book) ReviewString() string {
	return fmt.Sprintf("\"%v\" by %v %v", book.Title, utils.SliceList(book.Authors), strings.Repeat("*", book.Rating))
}

func (book *Book) SortedDate() time.Time {
	return book.DateAdded
}

func (client *Client) getBooks(userId int) (map[string]*Book, error) {
	bookMap := client.readBooksCache()
	bookMap = client.GetBooksRequest(userId, bookMap)
	return bookMap, client.saveBooksCache(bookMap)
}

type jsonBooksRoot struct {
	Books map[string]*Book `json:"books"`
}

func (client *Client) readBooksCache() map[string]*Book {
	jsonRoot := jsonBooksRoot{}
	booksCachePath := path.Join(client.Settings.CachePath, booksFile)

	_, err := os.Stat(booksCachePath)
	if os.IsNotExist(err) {
		client.log.Infof("%v does not exist - %v", booksCachePath, err)
		return nil
	}

	bytes, err := ioutil.ReadFile(booksCachePath)
	if err != nil {
		client.log.Warnf("error reading %v - %v", booksCachePath, err)
		return nil
	}

	err = json.Unmarshal(bytes, &jsonRoot)
	if err != nil {
		client.log.Warnf("error reading %v - %v", booksCachePath, err)
		return nil
	}

	if jsonRoot.Books == nil {
		client.log.Warnf("key books in %v is nil", booksCachePath)
	}
	client.log.Infof("Loaded %v books from %v", len(jsonRoot.Books), booksCachePath)
	return jsonRoot.Books
}

func (client *Client) saveBooksCache(bookMap map[string]*Book) error {
	bytes, err := json.MarshalIndent(jsonBooksRoot{bookMap}, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(client.Settings.CachePath, booksFile), bytes, 0755)
}

type xmlBookResponse struct {
	XMLName  xml.Name   `xml:"GoodreadsResponse"`
	PageData xmlReviews `xml:"reviews"`
}

func (response *xmlBookResponse) HasMore() bool {
	return response.PageData.PageEnd < response.PageData.TotalBooks
}

type xmlReviews struct {
	XMLName xml.Name `xml:"reviews"`

	Books []*Book `xml:"review"`

	PageStart  int `xml:"start,attr"`
	PageEnd    int `xml:"end,attr"`
	TotalBooks int `xml:"total,attr"`
}

func (client *Client) GetBooksRequest(userId int, bookMap map[string]*Book) map[string]*Book {
	if bookMap == nil {
		bookMap = map[string]*Book{}
	}

	initialLoad := len(bookMap) == 0
	if initialLoad {
		client.log.Info("Loading all data from goodreads API")
	}

	totalApiBooks := 0
	booksAdded := 0

	defer func() {
		if len(bookMap) < totalApiBooks {
			client.log.Warnf("bookMap has %v elements, while there are %v books in the API", len(bookMap), totalApiBooks)
		} else {
			client.log.Infof("bookMap has all %v books", totalApiBooks)
		}
	}()

	err := client.paginateGet(
		func(page int) (resp *http.Response, err error) {
			return client.requestGetBooks(userId, initialLoad, page)
		},
		func(bytes []byte) (bool, error) {
			bookResponse := xmlBookResponse{}
			err := xml.Unmarshal(bytes, &bookResponse)
			if err != nil {
				return false, err
			}

			totalApiBooks = bookResponse.PageData.TotalBooks

			for _, book := range bookResponse.PageData.Books {
				if len(bookMap) >= totalApiBooks {
					return false, nil
				}
				if _, contains := bookMap[book.Id]; !contains {
					booksAdded += 1
					client.log.Infof("%v. %v", booksAdded, book.ReviewString())
					book.convertDates()
					bookMap[book.Id] = book
				}
			}
			return bookResponse.HasMore() && len(bookMap) < totalApiBooks, nil
		},
	)
	if err != nil {
		client.log.Warnf("paginateGet error - %v", err)
	}
	return bookMap
}

func (client *Client) requestGetBooks(userId int, initialLoad bool, page int) (resp *http.Response, err error) {
	perPage := client.Settings.PerPage
	if initialLoad {
		perPage = client.Settings.MaxPerPage
	}

	queryParams := map[string]string{
		"v":  "2",
		"id": strconv.Itoa(userId),

		"key": client.Settings.ApiKey,

		"shelf": "read",

		"page":     strconv.Itoa(page),
		"per_page": strconv.Itoa(perPage),
		"sort":     "date_added",
		"order":    "d",
	}

	url := "https://www.goodreads.com/review/list?" + utils.ToSimpleQuery(queryParams)
	client.log.Infof("GET %v", url)
	return http.Get(url)
}

func (client *Client) paginateGet(request func(page int) (resp *http.Response, err error), callback func(bytes []byte) (bool, error)) error {
	rateLimit := time.Duration(client.Settings.RateLimit) * time.Millisecond
	ticker := time.NewTicker(rateLimit)
	defer ticker.Stop()

	page := 1
	hasMore := true
	for hasMore {
		response, err := request(page)
		if err != nil {
			return err
		}
		bytes, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return err
		}

		hasMore, err = callback(bytes)
		if err != nil {
			return err
		}
		if hasMore {
			page += 1
			client.log.Infof("Sleeping for %v...", rateLimit)
			<-ticker.C
		}
	}
	return nil
}
