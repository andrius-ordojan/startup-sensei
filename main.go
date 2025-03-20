package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gocolly/colly/v2"
)

// TODO: add rogue startup podcast
type episode struct {
	Title       string `selector:"h1.entry-title"`
	Content     string `selector:"div.entry-content"`
	PublishedAt string `selector:"time.entry-time"`
	Url         string
}
type podcast struct {
	domain          string
	archivePageLink string
	episodes        []*episode
}

func scrapeStartupsForTheRestOfUs() (episode, error) {
	podcast := podcast{
		domain:          "www.startupsfortherestofus.com",
		archivePageLink: "https://www.startupsfortherestofus.com/archives",
	}
	fmt.Println(podcast)

	return episode{}, nil
	const podcastlink = "https://www.startupsfortherestofus.com/archives"

	episodes := []*episode{}
	c := colly.NewCollector(
		colly.AllowedDomains("www.startupsfortherestofus.com"),
	)

	c.OnHTML("ul.archive-list a[href]", func(e *colly.HTMLElement) {
		c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	c.OnHTML(".content", func(e *colly.HTMLElement) {
		episode := &episode{}
		e.Unmarshal(episode)
		episode.Url = e.Request.URL.String()
		episodes = append(episodes, episode)
	})

	c.Visit(podcastlink)

	// TODO: write this to file instead
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(episodes)

	return episode{}, nil
}

func run() error {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	// defer cancel()
	// TODO: read from local file first so I can continue from where I left off

	scrapeStartupsForTheRestOfUs()
	return nil
}

// TODO: add graceful shutdown so it it doesn't ruin the datastrucuture when kill signal is sent
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
