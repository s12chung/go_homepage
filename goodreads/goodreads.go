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

	log "github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/utils"
)

const booksFile = "books.json"

type Client struct {
	Settings settings.GoodreadsSettings
}

func NewClient(settings settings.GoodreadsSettings) *Client {
	return &Client{settings}
}

func (client *Client) GetAll() error {
	if client.invalidSettings() {
		log.Warn("Invalid goodsreads Settings, skipping goodsreads API calls")
		return nil
	}
	err := client.setup()
	if err != nil {
		return err
	}

	books, err := client.GetBooks(client.Settings.UserId)
	if err != nil {
		return err
	}

	return client.saveBooksCache(books)
}

func (client *Client) invalidSettings() bool {
	return client.Settings.ApiKey == "" && client.Settings.UserId == 0
}

func (client *Client) setup() error {
	cachePath := client.Settings.CachePath
	return os.MkdirAll(cachePath, 0755)
}

type book struct {
	XMLName xml.Name `xml:"review" json:"-"`
	Id      string   `xml:"id" json:"id"`
	Name    string   `xml:"book>title" json:"name"`
	Authors []string `xml:"book>authors>author>name" json:"authors"`
	Isbn    string   `xml:"book>isbn" json:"isbn"`
	Isbn13  string   `xml:"book>isbn_13" json:"isbn_13"`
	Rating  int      `xml:"rating" json:"rating"`
}

func (book *book) ReviewString() string {
	return fmt.Sprintf("%v \"%v\" - %v", strings.Repeat("*", book.Rating), book.Name, strings.Join(book.Authors, ","))
}

func (client *Client) GetBooks(userId int) (map[string]book, error) {
	bookMap := client.readBooksCache()
	bookMap = client.GetBooksRequest(userId, bookMap)
	return bookMap, client.saveBooksCache(bookMap)
}

type jsonBooksRoot struct {
	Books map[string]book `json:"books"`
}

func (client *Client) readBooksCache() map[string]book {
	jsonRoot := jsonBooksRoot{}
	booksCachePath := path.Join(client.Settings.CachePath, booksFile)

	_, err := os.Stat(booksCachePath)
	if os.IsNotExist(err) {
		log.Infof("%v does not exist - %v", booksCachePath, err)
		return nil
	}

	bytes, err := ioutil.ReadFile(booksCachePath)
	if err != nil {
		log.Warnf("error reading %v - %v", booksCachePath, err)
		return nil
	}

	err = json.Unmarshal(bytes, &jsonRoot)
	if err != nil {
		log.Warnf("error reading %v - %v", booksCachePath, err)
		return nil
	}

	if jsonRoot.Books == nil {
		log.Warnf("key books in %v is nil", booksCachePath)
	}
	log.Infof("Loaded %v books from %v", len(jsonRoot.Books), booksCachePath)
	return jsonRoot.Books
}

func (client *Client) saveBooksCache(bookMap map[string]book) error {
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

	Books []book `xml:"review"`

	PageStart  int `xml:"start,attr"`
	PageEnd    int `xml:"end,attr"`
	TotalBooks int `xml:"total,attr"`
}

func (client *Client) GetBooksRequest(userId int, bookMap map[string]book) map[string]book {
	if bookMap == nil {
		bookMap = map[string]book{}
	}

	initialLoad := len(bookMap) == 0
	if initialLoad {
		log.Info("Loading all data from goodreads API")
	}

	totalApiBooks := 0
	booksAdded := 0

	defer func() {
		if len(bookMap) < totalApiBooks {
			log.Warnf("bookMap has %v elements, while there are %v books in the API", len(bookMap), totalApiBooks)
		} else {
			log.Infof("bookMap has all %v books", totalApiBooks)
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
					log.Infof("%v. %v", booksAdded, book.ReviewString())
					bookMap[book.Id] = book
				}
			}
			return bookResponse.HasMore() && len(bookMap) < totalApiBooks, nil
		},
	)
	if err != nil {
		log.Warnf("paginateGet error - %v", err)
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

		"page":     strconv.Itoa(page),
		"per_page": strconv.Itoa(perPage),
		"sort":     "date_added",
		"order":    "d",
	}

	url := "https://www.goodreads.com/review/list?" + utils.ToSimpleQuery(queryParams)
	log.Infof("GET %v", url)
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
		if err != nil {
			return err
		}

		hasMore, err = callback(bytes)
		if err != nil {
			return err
		}
		if hasMore {
			page += 1
			log.Infof("Sleeping for %v...", rateLimit)
			<-ticker.C
		}
	}
	return nil
}
