package goodreads

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/s12chung/go_homepage/utils"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
)

const cachePath = "cache"
const apiKey = ""
const userId = 0

func Get() error {
	books, err := GetBookReviews(userId)
	if err != nil {
		return err
	}

	return jsonCache(bookContainer{*books})
}

func jsonCache(v interface{}) error {
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
	Elements []book `json:"books"`
}

type book struct {
	XMLName xml.Name `xml:"review" json:"-"`
	Name    string   `xml:"book>title" json:"name"`
	Authors []string `xml:"book>authors>author>name" json:"authors"`
	Isbn    string   `xml:"book>isbn" json:"isbn"`
	Isbn13  string   `xml:"book>isbn_13" json:"isbn_13"`
	Rating  int      `xml:"rating" json:"rating"`
}

func GetBookReviews(userId int) (*[]book, error) {
	queryParams := map[string]string{
		"v":  "2",
		"id": strconv.Itoa(userId),

		"key": apiKey,

		"per_page": "200",
		"sort":     "date_added",
		"order":    "d",
	}

	response, err := http.Get("https://www.goodreads.com/review/list?" + utils.ToSimpleQuery(queryParams))

	if err != nil {
		return nil, err
	}

	bookContainer := bookContainer{}
	err = unmarshallElements(&bookContainer, response.Body)
	if err != nil {
		return nil, err
	}
	return &bookContainer.Elements, nil
}

func unmarshallElements(containerStruct interface{}, r io.Reader) error {
	elementsValue := reflect.ValueOf(containerStruct).Elem().FieldByName("Elements")
	if elementsValue.Kind() != reflect.Slice {
		return errors.New("containerStruct's Elements is not a slice")
	}

	elementType := elementsValue.Type().Elem()
	elementStructField, hasXMLName := elementType.FieldByName("XMLName")
	if !hasXMLName {
		return errors.New("containerStruct Elements does not have XMLName")
	}
	XMLName := elementStructField.Tag.Get("xml")
	if XMLName == "" {
		return errors.New("containerStruct Elements XMLName is empty")
	}

	decoder := xml.NewDecoder(r)
	for {
		token, err := decoder.Token()
		if token == nil {
			break
		}
		if err != nil {
			return err
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == XMLName {
				n := elementsValue.Len()
				elementsValue.Set(reflect.Append(elementsValue, reflect.Zero(elementType)))

				if err := decoder.DecodeElement(elementsValue.Index(n).Addr().Interface(), &element); err != nil {
					elementsValue.SetLen(n)
					return err
				}
			}
		}
	}
	return nil
}
