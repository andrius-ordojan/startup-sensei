package main

import (
	"encoding/json"
	"os"

	"github.com/gocolly/colly/v2"
)

// TODO: add rogue startup podcast
const podcastlink = "https://www.startupsfortherestofus.com/archives"

type episode struct {
	Title       string `selector:"h1.entry-title"`
	Content     string `selector:"div.entry-content"`
	PublishedAt string `selector:"time.entry-time"`
	Url         string
}

// TODO: add graceful shutdown so it it doesn't ruin the datastrucuture when kill signal is sent
func main() {
	// TODO: read from local file first so I can continue from where I left off
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
}
