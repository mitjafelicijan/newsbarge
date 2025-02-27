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

	"github.com/mmcdole/gofeed"
)

//go:embed template.html
var content embed.FS

type Item struct {
	Title    string
	Author   string
	Url      string
	UnixTime int64
	Content  string
	PubDate  string
	AudioUrl string
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

func (p *Payload) parseFeed(feedUrl string) error {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)

	if err != nil {
		log.Println(feedUrl, err)
		return err
	}

	log.Println(feedUrl)

	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			continue
		}

		newItem := Item{
			Author:  feed.Title,
			Title:   item.Title,
			Url:     item.Link,
			Content: item.Content,
			PubDate: item.Published,
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

		if isItemRecent(item.PublishedParsed, 7) {
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
	_ = outDirectory

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
		err = payload.parseFeed(feed)
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

	file, _ := content.ReadFile("template.html")
	tmpl, _ := template.New("").Parse(string(file))
	tmpl.Execute(outFile, payload)
}
