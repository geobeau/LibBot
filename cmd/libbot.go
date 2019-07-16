package main

import (
	"time"
	"log"
	"os"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"net/http"

	tb "gopkg.in/tucnak/telebot.v2"
)

type book struct {
	id string
	author string
	title string
}

func extractBooks(resp http.Response) []book {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
	  log.Fatal(err)
	}
	books := []book {}
	doc.Find("table.c tr:not(:first-child)").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		row := s.Find("td")
		id := row.Eq(0).Text()
		author := row.Eq(1).Text()
		selector := row.Eq(2).Find("font")
		for i := range selector.Nodes {
			selector.Eq(i).Remove()
		}
		title := row.Eq(2).Find("a[title]").Eq(0).Text()
		books = append(books, book{id, author, title})
	})
	return books
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

	b.Handle(tb.OnText, func(m *tb.Message) {
		log.Println("Received:", m.Text)
		query := m.Text
		books := searchBooks(query)
		for i := range books {
			log.Println(books[i])
			b.Send(m.Sender, books[i].author + " | " + books[i].title)
			if i >= 4 {
				break
			}
		}
		
	})
	log.Println("Handler started")
	b.Start()
}