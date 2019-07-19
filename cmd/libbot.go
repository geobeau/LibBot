package main

import (
	"time"
	"log"
	"os"
	"strings"
	"fmt"
	"mime"
	"regexp"
	"github.com/PuerkitoBio/goquery"
	"net/http"

	tb "gopkg.in/tucnak/telebot.v2"
)

type book struct {
	id string
	author string
	title string
	year string
	checksum string
	format string
	pages string
	size string
	publisher string
}

type extendedBook struct {
	book book
	language string
	isbn string
	coverURL string
}

func extractBooks(resp http.Response) []book {
	r, _ := regexp.Compile("md5=([a-zA-Z0-9]+)")

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
	  log.Fatal(err)
	}
	books := []book {}
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
		publisher := row.Eq(3).Text()
		books = append(books, book{id, author, title, year, checksum, format, pages, size, publisher})
	})
	return books
}

func extractBookInfo(resp http.Response) extendedBook {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
	  log.Fatal(err)
	}

	r, _ := regexp.Compile("libgen.me - (.*)")
	rawTitle := doc.Find("title").Eq(0).Text()
	match := r.FindStringSubmatch(rawTitle)
	title := ""
	if len(match) > 1 {
		title = match[1]
	}
		

	author := doc.Find(".book-info__lead").Eq(0).Text()

	selector := doc.Find(".book-info__params tr")
	format := selector.Eq(1).Find("td").Eq(1).Text()
	publisher := selector.Eq(3).Find("td").Eq(1).Text()
	language := selector.Eq(4).Find("td").Eq(1).Text()
	year := selector.Eq(5).Find("td").Eq(1).Text()
	isbn := selector.Eq(6).Find("td").Eq(1).Text()
	size := selector.Eq(10).Find("td").Eq(1).Text()
	pages := selector.Eq(16).Find("td").Eq(1).Text()

	coverURL := "https://libgen.me" + doc.Find(".book-cover img").Eq(0).AttrOr("src", "")
	log.Println(coverURL)
	bookData := book{"", author, title, year, "", format, pages, size, publisher}
	bookInfoData := extendedBook{bookData, language, isbn, coverURL}
	return bookInfoData
}

func extractDownloadURL(resp http.Response) string {
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

func getBookFile(checksum string) (*http.Response, error) {
	apiBaseURL := "http://93.174.95.29"
	apiURL := fmt.Sprintf(apiBaseURL + "/_ads/%s", checksum)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return nil, err
	}
	downloadURL := apiBaseURL + extractDownloadURL(*resp)
	log.Println("Downloading: ", downloadURL)
	resp, err = http.Get(downloadURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return nil, err
	}
	return resp, nil
}

func fetchBookInfo(id string) (extendedBook, error) {
	apiBaseURL := "https://libgen.me/item/detail/id/%s" 
	apiURL := fmt.Sprintf(apiBaseURL, id)
	log.Println("Fetching: ", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return extendedBook{}, err
	}

	extendedBookData := extractBookInfo(*resp)
	return extendedBookData, nil
}


func formatBookMessage(book book) string {
	template := 
		"*%s*\n" +
		"By _%s_\n" +
		"%s | %s"
	message := fmt.Sprintf(template, book.title, book.author, book.year, book.format)
	return message
}

func formatInfoBookMessage(extendedBookData extendedBook) string {
	template := 
		"Title: *%s*\n" +
		"Author: _%s_\n" +
		"Year: %s\n" +
		"Format: %s\n" +
		"Pages: %s\n" +
		"Publisher: %s\n" +
		"Language: %s\n" +
		"ISBN: %s\n"
	book := extendedBookData.book
	message := fmt.Sprintf(template, book.title, book.author, book.year, book.format,
		book.pages, book.publisher, extendedBookData.language, extendedBookData.isbn)
	return message
}

func searchBooks(query string) []book {
	// Todo: make it configurable
	cleanQuery := strings.Replace(query, " ", "+", -1)
	apiURL := "https://libgen.is/search.php?req=" + cleanQuery
	log.Printf(apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Println("Failed to query URL: ", apiURL)
		return []book {}
	}
	return extractBooks(*resp)
}

func main() {

	log.Println("Starting libbot")
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Token is not set, please set BOT_TOKEN env variable")
		return
	}
	
	b, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Connected to api")

	downloadButton := tb.InlineButton{
		Unique: "download_button",
		Text: "Download",
	}
	infoButton := tb.InlineButton{
		Unique: "info_button",
		Text: "More info",
	}

	b.Handle(&infoButton, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Fetching more data..."})
		log.Println("Fetching more details about: ", c.Data)
		extendedBookData, err := fetchBookInfo(c.Data)
		if err != nil {
			log.Println("Failed to query URL: ", err)
			return
		}
		message := formatInfoBookMessage(extendedBookData)
		p := &tb.Photo{File: tb.FromURL(extendedBookData.coverURL)}
		p.Caption = message
		b.Send(c.Sender, p, tb.ModeMarkdown)
	})

	b.Handle(&downloadButton, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Downloading..."})
		bookResp, _ := getBookFile(c.Data)
		_, params, _ := mime.ParseMediaType(bookResp.Header.Get("Content-Disposition"))
		telegramFile := tb.FromReader(bookResp.Body)
		telegramFile.FileName = params["filename"]
		bookFile := &tb.Document{File: telegramFile}
		log.Println("Sending: ", params["filename"])
		bookFile.Send(b, c.Sender, nil)
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		log.Println("Received:", m.Text)
		query := m.Text
		b.Send(m.Sender, "Searching...")
		books := searchBooks(query)
		if len(books) == 0 {
			b.Send(m.Sender, "No result found")
			return
		}
		for i := range books {
			log.Println(books[i])
			downloadButton.Data = books[i].checksum
			infoButton.Data = books[i].id
			inlineButtons := [][]tb.InlineButton{
				[]tb.InlineButton{infoButton, downloadButton},
			}
			b.Send(m.Sender, formatBookMessage(books[i]), tb.ModeMarkdown, &tb.ReplyMarkup{
				InlineKeyboard: inlineButtons,
			})
			if i >= 10 {
				break
			}
		}		
	})

	log.Println("Handler started")
	b.Start()
}