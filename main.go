package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

const podcastFilePath = "content.json"

type episode struct {
	Title       string
	PublishedAt string
	Content     string
	Url         string
}

type Podcast struct {
	Domain          string
	ArchivePageLink string
	Episodes        []*episode
	podcastFile     *os.File
	scrape          func(ctx context.Context)
}

func (p *Podcast) addEpisode(e episode) error {
	p.Episodes = append(p.Episodes, &e)

	if len(p.Episodes)%10 == 0 {
		log.Println("total episodes: ", len(p.Episodes))
		if err := p.encode(); err != nil {
			return fmt.Errorf("could not encode podcast: %w", err)
		}
	}

	return nil
}

func (p *Podcast) encode() error {
	if err := p.podcastFile.Truncate(0); err != nil {
		return err
	}
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return err
	}

	enc := json.NewEncoder(p.podcastFile)
	return enc.Encode(p)
}

func (p *Podcast) decode() error {
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek file: %v", err)
	}

	dec := json.NewDecoder(p.podcastFile)
	if err := dec.Decode(p); err != nil {
		return fmt.Errorf("cloud not decode JSON: %w", err)
	}
	return nil
}

func (p *Podcast) deleteTempFile() {
	if err := os.Remove(p.podcastFile.Name()); err != nil {
		fmt.Println("could not delete temp file: ", err)
	}
}

func newRogueStartupsPodcast() (*Podcast, error) {
	file, err := os.Create("roguestartups.json")
	if err != nil {
		return &Podcast{}, fmt.Errorf("could not create podcast file: %w", err)
	}

	p := &Podcast{
		Domain:          "roguestartups.com",
		ArchivePageLink: "https://roguestartups.com/?page=1",

		podcastFile: file,
	}
	p.scrape = func(ctx context.Context) {
		c := colly.NewCollector(colly.AllowedDomains(p.Domain))

		c.OnHTML("body", func(e *colly.HTMLElement) {
			if ctx.Err() != nil {
				// TODO: rewerite this
				log.Println("context cancelled in  body")
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

			episode := &episode{
				Title:       title,
				PublishedAt: publishedDate,
				Content:     fmt.Sprintf("Show Notes: %s Transcript: %s", showNotes, transcript),
				Url:         e.Request.URL.String(),
			}

			if err := p.addEpisode(*episode); err != nil {
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

		c.Visit(p.ArchivePageLink)
	}
	return p, nil
}

func NewStartupsForTheRestOfUsPodcast() (*Podcast, error) {
	file, err := os.Create("roguestartups.json")
	if err != nil {
		return &Podcast{}, fmt.Errorf("could not create podcast file: %w", err)
	}

	p := &Podcast{
		Domain:          "www.startupsfortherestofus.com",
		ArchivePageLink: "https://www.startupsfortherestofus.com/archives",
		podcastFile:     file,
	}
	p.scrape = func(ctx context.Context) {
		c := colly.NewCollector(colly.AllowedDomains(p.Domain))

		c.OnError(func(r *colly.Response, err error) {
			fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		})

		c.OnHTML("ul.archive-list a[href]", func(e *colly.HTMLElement) {
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
			e.DOM.Find(".social-share").Remove()
			e.DOM.Find(".podcast_player").Remove()

			title := strings.TrimSpace(e.DOM.Find("h1.entry-title").Text())
			publishedDate := strings.TrimSpace(e.DOM.Find("time.entry-time").Text())

			content := strings.TrimSpace(e.DOM.Find("div.entry-content").Text())
			re := regexp.MustCompile(`\s+`)
			content = re.ReplaceAllString(content, " ")

			episode := &episode{
				Title:       title,
				PublishedAt: publishedDate,
				Content:     content,
				Url:         e.Request.URL.String(),
			}

			p.addEpisode(*episode)
		})

		c.Visit(p.ArchivePageLink)
	}
	return p, nil
}

type podcasts struct {
	Podcasts    []*Podcast
	podcastFile *os.File
}

func NewPodcasts(podcastFile *os.File) (*podcasts, error) {
	p := &podcasts{
		podcastFile: podcastFile,
	}
	p.decode()

	if len(p.Podcasts) == 0 {
		pod, err := newRogueStartupsPodcast()
		if err != nil {
			return &podcasts{}, fmt.Errorf("could not create podcast: %w", err)
		}
		p.Podcasts = append(p.Podcasts, pod)

		pod, err = NewStartupsForTheRestOfUsPodcast()
		if err != nil {
			return &podcasts{}, fmt.Errorf("could not create podcast: %w", err)
		}
		p.Podcasts = append(p.Podcasts, pod)
	} else {
		s := fmt.Sprintf("decoded: %s - %d episodes; %s - %d episodes",
			p.Podcasts[0].Domain,
			len(p.Podcasts[0].Episodes),
			p.Podcasts[1].Domain,
			len(p.Podcasts[1].Episodes))
		log.Println(s)
	}

	return p, nil
}

func (p *podcasts) encode() error {
	if err := p.podcastFile.Truncate(0); err != nil {
		return err
	}
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return err
	}

	enc := json.NewEncoder(p.podcastFile)
	return enc.Encode(p)
}

func (p *podcasts) decode() error {
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek file: %v", err)
	}

	dec := json.NewDecoder(p.podcastFile)
	if err := dec.Decode(p); err != nil {
		return fmt.Errorf("cloud not decode JSON: %w", err)
	}
	return nil
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var podcastFile *os.File

	if _, err := os.Stat(podcastFilePath); err == nil {
		podcastFile, err = os.OpenFile(podcastFilePath, os.O_RDWR, 0666)
		if err != nil {
			return fmt.Errorf("could not open file: %w", err)
		}
	} else if os.IsNotExist(err) {
		podcastFile, err = os.Create(podcastFilePath)
		if err != nil {
			return fmt.Errorf("could not create podcast file: %w", err)
		}
	} else {
		return fmt.Errorf("could not stat file: %w", err)
	}
	defer podcastFile.Close()

	pods, err := NewPodcasts(podcastFile)
	if err != nil {
		return fmt.Errorf("could not create podcasts: %w", err)
	}

	for _, p := range pods.Podcasts {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received, aborting scraping")
			return ctx.Err()
		default:
			p.scrape(ctx)
		}
	}

	// BUG: doesn't save the pods
	log.Println("saving podcasts")
	pods.encode()
	for _, p := range pods.Podcasts {
		log.Println("deleting temp file")
		// BUG: doesn't delete the temp file
		p.deleteTempFile()
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
