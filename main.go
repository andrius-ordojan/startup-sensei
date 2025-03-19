package main

import (
	"github.com/gocolly/colly/v2"
)

const podcastlink = "https://www.startupsfortherestofus.com/archives"

type episode struct {
	Title       string `selector:"h1.entry-title"`
	Content     string `selector:"div.entry-content"`
	PublishedAt string `selector:"time.entry-time"`
	Url         string
}

func main() {
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
}
