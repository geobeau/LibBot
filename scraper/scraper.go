package scraper

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"fmt"

	"github.com/geobeau/Libbot/book"
	"github.com/PuerkitoBio/goquery"
)

// ExtractBookMetadata extracts metadata from a webpage
func ExtractBookMetadata(resp http.Response) book.Book {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	title := strings.TrimSpace(doc.Find(".itemFullText h1").Eq(0).Text())

	author := strings.TrimSpace(doc.Find(".itemFullText i a").Eq(0).Text())

	selector := doc.Find(".bookDetailsBox").Eq(0)

	// Remove all the spans to make it easy to fetch data
	selRm := selector.Find("span")
	for i := range selRm.Nodes {
		selRm.Eq(i).Remove()
	}
	year := strings.TrimSpace(selector.Find(".property_year").Eq(0).Text())
	language := strings.TrimSpace(selector.Find(".property_language").Eq(0).Text())
	pages := strings.TrimSpace(selector.Find(".property_pages").Eq(0).Text())
	isbn := strings.TrimSpace(selector.Find(".property_isbn").Eq(0).Text())
	fileData := strings.TrimSpace(selector.Find(".property__file").Eq(0).Text())
	files := strings.Split(fileData, ",")
	format := files[0]
	size := files[1]
	coverURL := "https:" + doc.Find(".cardBooks .details-book-cover img").Eq(0).AttrOr("src", "")
	bookMetadata := book.Book{"", author, title, year, "", format, pages, size, language, isbn, coverURL}
	return bookMetadata
}

// extractBooksFromList extracts multiple book's metada from a search web page
func extractBooksFromList(resp http.Response) []book.Book {
	r, _ := regexp.Compile("md5=([a-zA-Z0-9]+)")

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	books := []book.Book{}
	doc.Find("table.c tr:not(:first-child)").Each(func(i int, s *goquery.Selection) {
		row := s.Find("td")
		id := row.Eq(0).Text()
		author := row.Eq(1).Text()
		selector := row.Eq(2).Find("font")
		for i := range selector.Nodes {
			selector.Eq(i).Remove()
		}
		infoURL := row.Eq(2).Find("a[title]").Eq(0).AttrOr("href", "")
		// Todo: make it more robust
		checksum := r.FindStringSubmatch(infoURL)[1]
		title := row.Eq(2).Find("a[title]").Eq(0).Text()
		year := row.Eq(4).Text()
		format := row.Eq(8).Text()
		pages := row.Eq(5).Text()
		size := row.Eq(7).Text()
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
func FetchBookMetadata(checksum string) (book.Book, error) {
	apiBaseURL := "https://b-ok.cc"
	apiURL := fmt.Sprintf(apiBaseURL+"/md5/%s", checksum)
	log.Println("Fetching: ", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return book.Book{}, err
	}
	dataURL := ExtractDetailedMetadataURL(*resp)
	log.Println("Fetching: ", apiBaseURL+dataURL)
	resp, err = http.Get(apiBaseURL + dataURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return book.Book{}, err
	}
	bookMetadata := ExtractBookMetadata(*resp)
	bookMetadata.Checksum = checksum
	return bookMetadata, nil
}

// GetBookFile Download the book file
func GetBookFile(checksum string) (*http.Response, error) {
	apiBaseURL := "http://93.174.95.29"
	apiURL := fmt.Sprintf(apiBaseURL+"/_ads/%s", checksum)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return nil, err
	}
	downloadURL := apiBaseURL + ExtractDownloadURL(*resp)
	log.Println("Downloading: ", downloadURL)
	resp, err = http.Get(downloadURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return nil, err
	}
	return resp, nil
}

// searchBooks search for books
func SearchBooks(query string) []book.Book {
	// Todo: make it configurable
	cleanQuery := strings.Replace(query, " ", "+", -1)
	apiURL := "https://libgen.is/search.php?req=" + cleanQuery
	log.Printf(apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return []book.Book{}
	}
	return extractBooksFromList(*resp)
}