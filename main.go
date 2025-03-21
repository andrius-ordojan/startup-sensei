package main

import (
	"fmt"
	"os"

	"github.com/gocolly/colly/v2"
)

const resultFile = "result.json"

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

func scrapeStartupsForTheRestOfUs(resultFile *os.File) (episode, error) {
	podcast := podcast{
		domain:          "www.startupsfortherestofus.com",
		archivePageLink: "https://www.startupsfortherestofus.com/archives",
	}
	fmt.Println(pod)

	c := colly.NewCollector(colly.AllowedDomains(pod.domain))

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.OnHTML("ul.archive-list a[href]", func(e *colly.HTMLElement) {
		c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	c.OnHTML(".content", func(e *colly.HTMLElement) {
		episode := &episode{}
		e.Unmarshal(episode)
		episode.Url = e.Request.URL.String()
		pod.episodes = append(pod.episodes, episode)
	})

	c.Visit(pod.archivePageLink)

	return pod, nil

	// // TODO: write this to file instead
	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "  ")
	// enc.Encode(episodes)
	//
	// return episode{}, nil
}

func run() error {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	// defer cancel()
	// TODO: read from local file first so I can continue from where I left off

	var file *os.File
	defer file.Close()
	if _, err := os.Stat(resultFile); err == nil {
		// TODO: open the file
		// read the data to pass to scraper
	} else if os.IsNotExist(err) {
		file, err = os.Create(resultFile)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return err
		}
	}
	scrapeStartupsForTheRestOfUs(file)

	return nil
}

// TODO: add graceful shutdown so it it doesn't ruin the datastrucuture when kill signal is sent
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
