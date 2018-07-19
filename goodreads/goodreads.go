package goodreads

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/utils"
)

type Client struct {
	Settings    settings.GoodreadsSettings
	InitialLoad bool
}

func NewClient(settings settings.GoodreadsSettings, initialLoad bool) *Client {
	return &Client{
		settings,
		initialLoad,
	}
}

func (client *Client) GetAll() error {
	if client.invalidSettings() {
		return nil
	}

	books, err := client.GetBooks(client.Settings.UserId)
	if err != nil {
		return err
	}

	return client.jsonCache(bookContainer{*books})
}

func (client *Client) invalidSettings() bool {
	return client.Settings.ApiKey == "" && client.Settings.UserId == 0
}

func (client *Client) jsonCache(v interface{}) error {
	cachePath := client.Settings.CachePath

	err := os.MkdirAll(cachePath, 0755)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(cachePath, "books.json"), bytes, 0755)
}

type bookContainer struct {
	Books []book `json:"books"`
}

type book struct {
	XMLName xml.Name `xml:"review" json:"-"`
	Name    string   `xml:"book>title" json:"name"`
	Authors []string `xml:"book>authors>author>name" json:"authors"`
	Isbn    string   `xml:"book>isbn" json:"isbn"`
	Isbn13  string   `xml:"book>isbn_13" json:"isbn_13"`
	Rating  int      `xml:"rating" json:"rating"`
}

func (client *Client) GetBooks(userId int) (*[]book, error) {
	response, err := client.requestGetBooks(userId)
	if err != nil {
		return nil, err
	}

	var books []book

	resultChan := decodeTag("review", response.Body)

	for result := range resultChan {
		if result.Error != nil {
			return nil, err
		}
		book := book{}
		result.Decode(&book)
		books = append(books, book)
	}
	return &books, nil
}

func (client *Client) requestGetBooks(userId int) (resp *http.Response, err error) {
	perPage := "50"
	if client.InitialLoad {
		perPage = "200"
	}

	queryParams := map[string]string{
		"v":  "2",
		"id": strconv.Itoa(userId),

		"key": client.Settings.ApiKey,

		"per_page": perPage,
		"sort":     "date_added",
		"order":    "d",
	}

	return http.Get("https://www.goodreads.com/review/list?" + utils.ToSimpleQuery(queryParams))
}

type decodeTagResult struct {
	Decode func(interface{})
	Error  error
}

func decodeTag(tag string, r io.Reader) <-chan decodeTagResult {
	resultChan := make(chan decodeTagResult)
	go func() {
		defer close(resultChan)

		decoder := xml.NewDecoder(r)
		for {
			token, err := decoder.Token()
			if token == nil {
				break
			}
			if err != nil {
				resultChan <- decodeTagResult{nil, err}
				break
			}

			switch element := token.(type) {
			case xml.StartElement:
				if element.Name.Local == tag {
					resultChan <- decodeTagResult{
						func(elementInterface interface{}) {
							decoder.DecodeElement(elementInterface, &element)
						},
						nil,
					}
				}
			}
		}
	}()
	return resultChan
}
