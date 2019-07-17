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
		books = append(books, book{id, author, title, year, checksum, format, pages})
	})
	return books
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


func formatBookMessage(book book) string {
	template := 
		"*%s*\n" +
		"By _%s_\n" +
		"%s | %s"
	message := fmt.Sprintf(template, book.title, book.author, book.year, book.format)
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
			inlineButtons := [][]tb.InlineButton{
				[]tb.InlineButton{downloadButton},
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