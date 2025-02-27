package main

import (
	"bufio"
	"embed"
	"flag"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"crypto/md5"
	"encoding/hex"

	"github.com/mmcdole/gofeed"
)

//go:embed templates/simple.html
//go:embed templates/advanced.html
var content embed.FS

type Item struct {
	Title      string
	Author     string
	Url        string
	UnixTime   int64
	Content    string
	PubDate    string
	PubDateRaw time.Time
	AudioUrl   string
	Hash string
}

type Payload struct {
	Items []Item
	Feeds []string
}

func isValidURL(str string) bool {
	_, err := url.ParseRequestURI(str)
	return err == nil
}

func isItemRecent(date *time.Time, days int) bool {
	if date == nil {
		return false
	}

	now := time.Now()
	cutoffTime := now.AddDate(0, 0, -days)
	return date.After(cutoffTime)
}

func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func isAudioType(mimeType string) bool {
	audioTypes := []string{
		"audio/mpeg", "audio/mp3", "audio/mp4", "audio/x-m4a",
		"audio/ogg", "audio/vorbis", "audio/aac", "audio/wav",
		"audio/webm", "audio/flac",
	}

	for _, t := range audioTypes {
		if t == mimeType {
			return true
		}
	}

	return false
}

func (p *Payload) parseFeed(feedUrl string, daysSpan int) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)

	if err != nil {
		log.Println(feedUrl, err)
		return err
	}

	log.Printf("parse %s", feedUrl)

	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		newItem := Item{
			Author:     feed.Title,
			Title:      item.Title,
			Url:        item.Link,
			Content:    item.Content,
			PubDate:    item.Published,
			PubDateRaw: *item.PublishedParsed,
			Hash: md5Hash(item.Link),
		}

		if itunesExt := feed.ITunesExt; itunesExt != nil {
			newItem.Content = itunesExt.Summary
			if len(item.Enclosures) > 0 {
				for _, enclosure := range item.Enclosures {
					if isAudioType(enclosure.Type) {
						newItem.AudioUrl = enclosure.URL
					}
				}
			}
		}

		if isItemRecent(item.PublishedParsed, daysSpan) {
			p.Items = append(p.Items, newItem)
		}
	}

	return nil
}

func (p *Payload) readFeedFile(feedFile string) error {
	file, err := os.Open(feedFile)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if len(trimmed) == 0 {
			continue
		}

		if isValidURL(line) {
			p.Feeds = append(p.Feeds, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	var err error

	feedFile := flag.String("feed-file", "", "URL of the file to download")
	outDirectory := flag.String("out-dir", ".", "Directory to save the downloaded file")
	daysSpan := flag.Int("days-span", 7, "Number of days")

	flag.Parse()

	if *feedFile == "" {
		log.Println("Error: -feed-file is required")
		flag.Usage()
		os.Exit(1)
	}

	payload := &Payload{}

	err = payload.readFeedFile(*feedFile)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	for _, feed := range payload.Feeds {
		err = payload.parseFeed(feed, *daysSpan)
		if err != nil {
			log.Println(err)
			panic(err)
		}
	}

	outPath := filepath.Join(*outDirectory, "newsbarge-recent.html")
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	file, _ := content.ReadFile("templates/advanced.html")
	// tmpl, _ := template.New("").Parse(string(file))
	tmpl, _ := template.New("").Funcs(funcMap).Parse(string(file))
	tmpl.Execute(outFile, payload)
	log.Printf("export file://%s", outPath)
}
