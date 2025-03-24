package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type Podcast interface {
	GetDomain() string
	ArchivePageLink() string
	Scrape(ctx context.Context)
	AddEpisode(e Episode) error
	GetEpisodes() []*Episode
	PodcastFile() *os.File
	DeletePodcastFile()
	Encode() error
}

type Episode struct {
	Title       string
	PublishedAt string
	Content     string
	Url         string
}

type RogueStartups struct {
	Domain      string
	Episodes    []*Episode
	podcastFile *os.File
}

func (p *RogueStartups) GetDomain() string {
	return p.Domain
}

func (p *RogueStartups) ArchivePageLink() string {
	return "https://roguestartups.com/?page=1"
}

func (p *RogueStartups) Scrape(ctx context.Context) {
	log.Println("scraping RogueStartups...")

	c := colly.NewCollector(colly.AllowedDomains(p.Domain))

	c.OnHTML("body", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			return
		}
		if !strings.Contains(e.Request.URL.Path, "/episodes/") {
			return
		}

		title := strings.TrimSpace(e.DOM.Find("h1.text-3xl").Text())
		publishedDate := strings.TrimSpace(e.DOM.Find("span.text-sm.text-skin-a11y").First().Text())

		re := regexp.MustCompile(`\s+`)
		transcript := re.ReplaceAllString(strings.TrimSpace(e.DOM.Find("#transcript-body").Text()), " ")
		showNotes := ""
		e.DOM.Find("h2").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "Show Notes") {
				showNotes = re.ReplaceAllString(strings.TrimSpace(s.Parent().Find("div.prose").Text()), " ")
			}
		})

		episode := &Episode{
			Title:       title,
			PublishedAt: publishedDate,
			Content:     fmt.Sprintf("Show Notes: %s Transcript: %s", showNotes, transcript),
			Url:         e.Request.URL.String(),
		}

		if err := p.AddEpisode(*episode); err != nil {
			log.Printf("error adding episode: %v", err)
		}
	})

	c.OnHTML("a[href^='/episodes/']", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			return
		}

		episodeURL := e.Request.AbsoluteURL(e.Attr("href"))

		exists := false
		for _, episode := range p.Episodes {
			if episode.Url == episodeURL {
				exists = true
				break
			}
		}
		if !exists {
			e.Request.Visit(episodeURL)
		}
	})

	c.OnHTML("a[href*='?page=']", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			return
		}

		text := strings.TrimSpace(e.Text)
		if text == "Next" {
			nextPageURL := e.Request.AbsoluteURL(e.Attr("href"))
			e.Request.Visit(nextPageURL)
		}
	})

	c.Visit(p.ArchivePageLink())
}

func (p *RogueStartups) AddEpisode(e Episode) error {
	p.Episodes = append(p.Episodes, &e)

	if len(p.Episodes)%10 == 0 {
		log.Println("total episodes: ", len(p.Episodes))
		if err := p.Encode(); err != nil {
			return fmt.Errorf("could not encode podcast: %w", err)
		}
	}

	return nil
}

func (p *RogueStartups) GetEpisodes() []*Episode {
	return p.Episodes
}

func (p *RogueStartups) Encode() error {
	return encode(p)
}

func (p *RogueStartups) PodcastFile() *os.File {
	return p.podcastFile
}

func (p *RogueStartups) DeletePodcastFile() {
	p.podcastFile.Close()
	deleteFile(p.podcastFile.Name())
}

func newRogueStartupsPodcast() (Podcast, error) {
	file, err := createFile("roguestartups.json")
	if err != nil {
		return &RogueStartups{}, fmt.Errorf("could not create podcast file: %w", err)
	}

	p := &RogueStartups{
		Domain:      "roguestartups.com",
		Episodes:    []*Episode{},
		podcastFile: file,
	}

	return p, nil
}

type StartupsForTheRestOfUs struct {
	Domain      string
	Episodes    []*Episode
	podcastFile *os.File
}

func (p *StartupsForTheRestOfUs) GetDomain() string {
	return p.Domain
}

func (p *StartupsForTheRestOfUs) ArchivePageLink() string {
	return "https://www.startupsfortherestofus.com/archives"
}

func (p *StartupsForTheRestOfUs) Scrape(ctx context.Context) {
	log.Println("scraping StartupsForTheRestOfUs...")

	c := colly.NewCollector(colly.AllowedDomains(p.Domain))

	c.OnHTML("ul.archive-list a[href]", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			return
		}

		exists := false
		for _, episode := range p.Episodes {
			if episode.Url == e.Request.AbsoluteURL(e.Attr("href")) {
				exists = true
				break
			}
		}
		if !exists {
			c.Visit(e.Request.AbsoluteURL(e.Attr("href")))
		}
	})

	c.OnHTML(".content", func(e *colly.HTMLElement) {
		if ctx.Err() != nil {
			return
		}

		e.DOM.Find(".social-share").Remove()
		e.DOM.Find(".podcast_player").Remove()

		title := strings.TrimSpace(e.DOM.Find("h1.entry-title").Text())
		publishedDate := strings.TrimSpace(e.DOM.Find("time.entry-time").Text())

		content := strings.TrimSpace(e.DOM.Find("div.entry-content").Text())
		re := regexp.MustCompile(`\s+`)
		content = re.ReplaceAllString(content, " ")

		episode := &Episode{
			Title:       title,
			PublishedAt: publishedDate,
			Content:     content,
			Url:         e.Request.URL.String(),
		}

		if err := p.AddEpisode(*episode); err != nil {
			log.Printf("error adding episode: %v", err)
		}
	})

	c.Visit(p.ArchivePageLink())
}

func (p *StartupsForTheRestOfUs) AddEpisode(e Episode) error {
	p.Episodes = append(p.Episodes, &e)

	if len(p.Episodes)%10 == 0 {
		log.Println("total episodes: ", len(p.Episodes))
		if err := p.Encode(); err != nil {
			return fmt.Errorf("could not encode podcast: %w", err)
		}
	}

	return nil
}

func (p *StartupsForTheRestOfUs) GetEpisodes() []*Episode {
	return p.Episodes
}

func (p *StartupsForTheRestOfUs) Encode() error {
	return encode(p)
}

func (p *StartupsForTheRestOfUs) PodcastFile() *os.File {
	return p.podcastFile
}

func (p *StartupsForTheRestOfUs) DeletePodcastFile() {
	p.podcastFile.Close()
	deleteFile(p.podcastFile.Name())
}

func NewStartupsForTheRestOfUsPodcast() (Podcast, error) {
	file, err := createFile("startupsfortherestofus.json")
	if err != nil {
		return &StartupsForTheRestOfUs{}, fmt.Errorf("could not create podcast file: %w", err)
	}

	p := &StartupsForTheRestOfUs{
		Domain:      "www.startupsfortherestofus.com",
		Episodes:    []*Episode{},
		podcastFile: file,
	}
	return p, nil
}

type Podcasts struct {
	Podcasts    []Podcast
	podcastFile *os.File
}

func (p *Podcasts) encode() error {
	defer p.podcastFile.Close()
	if err := p.podcastFile.Truncate(0); err != nil {
		return err
	}
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return err
	}

	enc := json.NewEncoder(p.podcastFile)
	return enc.Encode(p)
}

func (p *Podcasts) decode() error {
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek file: %v", err)
	}

	dec := json.NewDecoder(p.podcastFile)
	if err := dec.Decode(p); err != nil {
		return fmt.Errorf("cloud not decode JSON: %w", err)
	}
	return nil
}

func NewPodcasts() (*Podcasts, error) {
	file, err := createFile("podcasts.json")
	if err != nil {
		return &Podcasts{}, fmt.Errorf("could not create podcast file: %w", err)
	}

	p := &Podcasts{
		podcastFile: file,
	}

	pod, err := newRogueStartupsPodcast()
	if err != nil {
		return &Podcasts{}, fmt.Errorf("could not create podcast: %w", err)
	}
	p.Podcasts = append(p.Podcasts, pod)

	pod, err = NewStartupsForTheRestOfUsPodcast()
	if err != nil {
		return &Podcasts{}, fmt.Errorf("could not create podcast: %w", err)
	}
	p.Podcasts = append(p.Podcasts, pod)

	p.decode()

	if len(p.Podcasts) > 0 {
		s := fmt.Sprintf("decoded: %s - %d episodes; %s - %d episodes",
			p.Podcasts[0].GetDomain(),
			len(p.Podcasts[0].GetEpisodes()),
			p.Podcasts[1].GetDomain(),
			len(p.Podcasts[1].GetEpisodes()))
		log.Println(s)
	}

	return p, nil
}
