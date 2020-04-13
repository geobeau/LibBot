package scraper

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/geobeau/Libbot/book"
)

// ExtractBookMetadata extracts metadata from a webpage
func ExtractBookMetadata(resp http.Response) book.Book {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	id := doc.Find("a.dlButton").Eq(0).AttrOr("href", "")

	title := strings.TrimSpace(doc.Find(".itemFullText h1").Eq(0).Text())

	author := strings.TrimSpace(doc.Find(".itemFullText i a").Eq(0).Text())

	selector := doc.Find(".bookDetailsBox")

	year := strings.TrimSpace(selector.Find(".property_year .property_value").Eq(0).Text())
	language := strings.TrimSpace(selector.Find(".property_language .property_value").Eq(0).Text())
	pages := selector.Find(".property_pages span").Eq(0).Text()
	isbn := strings.TrimSpace(selector.Find(".property_isbn .property_value").Eq(0).Text())
	fileData := strings.TrimSpace(selector.Find(".property__file .property_value").Eq(0).Text())
	files := strings.Split(fileData, ",")
	format := files[0]
	size := files[1]
	coverURL := doc.Find(".cardBooks .details-book-cover img").Eq(0).AttrOr("src", "")
	bookMetadata := book.Book{id, author, title, year, "", format, pages, size, language, isbn, coverURL}
	return bookMetadata
}

// extractBooksFromList extracts multiple book's metada from a search web page
func extractBooksFromList(resp http.Response) []book.Book {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	books := []book.Book{}
	doc.Find("div#searchResultBox div.resItemBox").Each(func(i int, s *goquery.Selection) {
		id := s.Find("h3 a").Eq(0).AttrOr("href", "")
		authors := []string{}
		selector := s.Find(".authors a")
		for i := range selector.Nodes {
			authors = append(authors, selector.Eq(i).Text())
		}
		author := strings.Join(authors, "")

		checksum := ""
		title := s.Find("h3 a").Eq(0).Text()
		year := s.Find(".property_year .property_value").Eq(0).Text()
		file := strings.Split(s.Find(".property__file .property_value").Eq(0).Text(), ",")
		format := file[0]
		pages := ""
		size := file[1]
		books = append(books, book.Book{id, author, title, year, checksum, format, pages, size, "", "", ""})
	})
	return books
}

// ExtractDownloadURL extracts the URL to download a book from a webpage
func ExtractDownloadURL(resp http.Response) string {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	downloadURL := ""
	log.Println("Searching")
	doc.Find("#info a").Each(func(i int, s *goquery.Selection) {
		if s.Text() == "GET" {
			downloadURL = s.AttrOr("href", "")
		}
	})
	return downloadURL
}

// ExtractDetailedMetadataURL extracts metadata URL which is used to get more informations about a book
func ExtractDetailedMetadataURL(resp http.Response) string {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Searching")
	return doc.Find("td.itemCover a").Eq(0).AttrOr("href", "")
}

// FetchBookMetadata crawl and parse the correct api to fetch book metadata
func FetchBookMetadata(id string) (book.Book, error) {
	apiBaseURL := "https://b-ok.cc%s"
	apiURL := fmt.Sprintf(apiBaseURL, id)
	log.Println(apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return book.Book{}, err
	}
	bookMetadata := ExtractBookMetadata(*resp)
	bookMetadata.Checksum = id
	return bookMetadata, nil
}

// GetBookFile Download the book file
func GetBookFile(id string) (*http.Response, error) {
	apiBaseURL := "https://b-ok.cc%s"
	downloadURL := fmt.Sprintf(apiBaseURL, id)
	log.Println("Downloading: ", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Println("Failed to query URL: ", downloadURL)
		return nil, err
	}
	return resp, nil
}

// SearchBooks search for books
func SearchBooks(query string) []book.Book {
	cleanQuery := url.QueryEscape(query)
	apiURL := "https://b-ok.cc/s/" + cleanQuery
	log.Printf(apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return []book.Book{}
	}
	return extractBooksFromList(*resp)

}
