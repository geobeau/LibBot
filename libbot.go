package main

import (
	"fmt"
	"log"
	"mime"
	"os"
	"time"
	"github.com/geobeau/Libbot/scraper"
	"github.com/geobeau/Libbot/book"
	tb "gopkg.in/tucnak/telebot.v2"
)

func formatBookMessage(book book.Book) string {
	template :=
		"*%s*\n" +
			"By _%s_\n" +
			"%s | %s"
	message := fmt.Sprintf(template, book.Title, book.Author, book.Year, book.Format)
	return message
}

func formatInfoBookMessage(book book.Book) string {
	template :=
		"Title: *%s*\n" +
			"Author: _%s_\n" +
			"Year: %s\n" +
			"Format: %s\n" +
			"Pages: %s\n" +
			"Language: %s\n" +
			"ISBN: %s\n"
	message := fmt.Sprintf(template, book.Title, book.Author, book.Year, book.Format,
		book.Pages, book.Language, book.Isbn)
	return message
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
		Text:   "Download",
	}
	infoButton := tb.InlineButton{
		Unique: "info_button",
		Text:   "More info",
	}

	b.Handle(&infoButton, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Fetching more data..."})
		log.Println("Fetching more details about: ", c.Data)
		bookMetadata, err := scraper.FetchBookMetadata(c.Data)
		if err != nil {
			log.Println("Failed to query URL: ", err)
			return
		}
		message := formatInfoBookMessage(bookMetadata)
		log.Println(bookMetadata.CoverURL, message)
		p := &tb.Photo{File: tb.FromURL(bookMetadata.CoverURL)}
		p.Caption = message
		downloadButton.Data = bookMetadata.Checksum
		inlineButtons := [][]tb.InlineButton{
			[]tb.InlineButton{downloadButton},
		}
		b.Send(c.Sender, p, tb.ModeMarkdown, &tb.ReplyMarkup{
			InlineKeyboard: inlineButtons,
		})
	})

	b.Handle(&downloadButton, func(c *tb.Callback) {
		b.Respond(c, &tb.CallbackResponse{Text: "Downloading..."})
		bookResp, _ := scraper.GetBookFile(c.Data)
		_, params, _ := mime.ParseMediaType(bookResp.Header.Get("Content-Disposition"))
		telegramFile := tb.FromReader(bookResp.Body)
		telegramFile.FileName = params["filename"]
		bookFile := &tb.Document{File: telegramFile}
		log.Println("Sending: ", params["filename"])
		_, err := bookFile.Send(b, c.Sender, nil)
		if err != nil {
			log.Println("Error:", err)
		}
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		log.Println("Received:", m.Text)
		query := m.Text
		b.Send(m.Sender, "Searching...")
		books := scraper.SearchBooks(query)
		if len(books) == 0 {
			b.Send(m.Sender, "No result found")
			return
		}
		for i := range books {
			log.Println(books[i])
			downloadButton.Data = books[i].Checksum
			infoButton.Data = books[i].Checksum
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
