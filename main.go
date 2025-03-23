package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gocolly/colly/v2"
)

const podcastFilePath = "content.json"

// TODO: add rogue startup podcast
type episode struct {
	Title       string `selector:"h1.entry-title"`
	Content     string `selector:"div.entry-content"`
	PublishedAt string `selector:"time.entry-time"`
	Url         string
}
type podcast struct {
	Domain          string
	ArchivePageLink string
	Episodes        []*episode
	podcastFile     *os.File
}

func (p *podcast) addEpisode(e episode) error {
	p.Episodes = append(p.Episodes, &e)

	if len(p.Episodes)%10 == 0 {
		fmt.Println("total episodes: ", len(p.Episodes))
		if err := p.encode(); err != nil {
			return fmt.Errorf("could not encode podcast: %w", err)
		}
	}

	return nil
}

func (p *podcast) encode() error {
	if err := p.podcastFile.Truncate(0); err != nil {
		return err
	}
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return err
	}

	enc := json.NewEncoder(p.podcastFile)
	return enc.Encode(p)
}

func (p *podcast) decode() error {
	if _, err := p.podcastFile.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek file: %v", err)
	}

	dec := json.NewDecoder(p.podcastFile)
	if err := dec.Decode(p); err != nil {
		return fmt.Errorf("cloud not decode JSON: %w", err)
	}
	return nil
}

func NewPodcast(podcastFile *os.File) *podcast {
	p := &podcast{
		Domain:          "www.startupsfortherestofus.com",
		ArchivePageLink: "https://www.startupsfortherestofus.com/archives",
		podcastFile:     podcastFile,
	}
	// p.decode()
	return p
}

type podcasts struct {
	Podcasts    []*podcast
	podcastFile *os.File
}

func NewPodcasts(podcastFile *os.File) *podcasts {
	p := &podcasts{
		podcastFile: podcastFile,
	}
	p.decode()

	if len(p.Podcasts) == 0 {
		p.Podcasts = append(p.Podcasts,
			&podcast{
				Domain:          "www.startupsfortherestofus.com",
				ArchivePageLink: "https://www.startupsfortherestofus.com/archives",
				podcastFile:     podcastFile,
			})
		// TODO: rogue startups
		p.Podcasts = append(p.Podcasts,
			&podcast{
				Domain:          "www.startupsfortherestofus.com",
				ArchivePageLink: "https://www.startupsfortherestofus.com/archives",
				podcastFile:     podcastFile,
			})
	} else {
		s := fmt.Sprintf("decoded: %s - %d episodes; %s - %d episodes",
			p.Podcasts[0].Domain,
			len(p.Podcasts[0].Episodes),
			p.Podcasts[1].Domain,
			len(p.Podcasts[1].Episodes))
		fmt.Println(s)
	}

	return p
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

func scrapeRogueStartups(p *podcast) {
}

func scrapeStartupsForTheRestOfUs(p *podcast) {
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
		episode := &episode{}
		e.Unmarshal(episode)
		episode.Url = e.Request.URL.String()
		p.addEpisode(*episode)
	})

	c.Visit(p.ArchivePageLink)
}

func run() error {
	var podcastFile *os.File
	defer podcastFile.Close()

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
	}

	pods := NewPodcasts(podcastFile)
	fmt.Println(pods)
	// pod := NewPodcast(podcastFile)
	// scrapeStartupsForTheRestOfUs(pod)
	return nil
}

// TODO: add graceful shutdown so it it doesn't ruin the datastrucuture when kill signal is sent
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
