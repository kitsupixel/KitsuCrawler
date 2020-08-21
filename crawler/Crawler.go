package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Crawler provides a datastucture with settings and methods to crawl a website
type Crawler struct {
	domain        *url.URL
	MaxDepthLevel int
	IgnoreRobots  bool
	robots        *Robots
	userAgent     string
}

type invalidURLError struct{}

func (e invalidURLError) Error() string {
	return "URL is not valid"
}

// NewCrawler creates a default Crawler object
func NewCrawler(domain string, userAgent string) (*Crawler, error) {

	domain = strings.TrimRight(domain, "/")

	u, err := url.Parse(domain)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, &invalidURLError{}
	}

	robot := NewRobot()
	robot.Build(domain, userAgent)

	return &Crawler{
		domain:        u,
		MaxDepthLevel: 5,
		IgnoreRobots:  false,
		robots:        robot,
		userAgent:     userAgent,
	}, nil
}

// Start initiliazes the crawl of the domain
func (crawler *Crawler) Start() {
	domain, domainName := getDomain(crawler.domain)

	c := colly.NewCollector(
		colly.Async(),
		colly.MaxDepth(crawler.MaxDepthLevel),
		colly.UserAgent(crawler.userAgent),
	)

	randomDelay := 0 * time.Second
	if !crawler.IgnoreRobots {
		randomDelay = crawler.robots.getCrawlDelay()
	}

	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*" + domainName + ".*",
		Parallelism: runtime.NumCPU(),
		RandomDelay: randomDelay,
	})

	fileName := "outputs/" + domainName + ".txt"

	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))

		link = strings.TrimRight(link, "/")

		parts := strings.Split(link, "/")
		isFile := false
		if strings.Contains(parts[len(parts)-1], ".") {
			isFile = true
		}

		if !isFile && (!crawler.IgnoreRobots && crawler.robots.IsAllowed(link)) &&
			strings.Contains(link, domain) && strings.HasPrefix(link, "http") {
			// Valid link to crawl
			_, err := f.WriteString(link + "\n")
			if err != nil {
				fmt.Println(err)
				f.Close()
				return
			}

			e.Request.Visit(link)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(crawler.domain.String())
	// Wait until threads are finished
	c.Wait()

	f.Close()

	fmt.Print("\nRemoving duplicate links from file...")
	removeDuplicatesFromFile(fileName)
	fmt.Print("Done!")
}

func getDomain(u *url.URL) (string, string) {
	domain := ""
	domainName := ""

	parts := strings.Split(u.Host, ".")

	// Removes "www"
	if parts[0] == "www" {
		parts = parts[1:]
	}

	numberOfParts := len(parts)

	if numberOfParts >= 2 {
		if numberOfParts == 3 {
			domain = parts[numberOfParts-3] + "." + parts[numberOfParts-2] + "." + parts[numberOfParts-1]
			domainName = parts[numberOfParts-3] + "." + parts[numberOfParts-2]
			if numberOfParts == 3 && len(parts[numberOfParts-2]) <= 3 {
				domainName = parts[numberOfParts-3]
			}
		} else {
			domain = parts[numberOfParts-2] + "." + parts[numberOfParts-1]
			domainName = parts[numberOfParts-2]
		}
	}

	return domain, domainName
}

func removeDuplicatesFromFile(filePath string) {
	// read the lines
	line, _ := ioutil.ReadFile(filePath)
	// turn the byte slice into string format
	strLine := string(line)
	// split the lines by a space, can also change this
	lines := strings.Split(strLine, "\n")
	// remove the duplicates from lines slice (from func we created)
	removeDuplicates(&lines)
	// get the actual file
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	// err check
	if err != nil {
		log.Println(err)
	}
	// delete old one
	os.Remove(filePath)
	// create it again
	os.Create(filePath)
	// go through your lines
	for e := range lines {
		// write to the file without the duplicates
		f.Write([]byte(lines[e] + "\n")) // added a space here, but you can change this
	}
	// close file
	f.Close()
}

func removeDuplicates(lines *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *lines {
		if !found[x] {
			found[x] = true
			(*lines)[j] = (*lines)[i]
			j++
		}
	}
	*lines = (*lines)[:j]
}
