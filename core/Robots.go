package core

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type rule struct {
	isAllowed bool
	path      string
	pattern   *regexp.Regexp
}

type group struct {
	userAgent  string
	rules      []*rule
	crawlDelay time.Duration
}

// Robots is a datastructure that contains the Robots.txt information
type Robots struct {
	domain    *url.URL
	groups    map[string]*group
	sitemaps  []string
	host      string
	userAgent string
}

type invalidHostError struct{}

func (e invalidHostError) Error() string {
	return "URL is not valid for this robots.txt file"
}

// NewRobot create a default Robot object
func NewRobot() *Robots {
	return &Robots{
		groups: make(map[string]*group),
	}
}

// Build gets the robots.txt file and stores it's information on the Robot object
func (robot *Robots) Build(domain string, userAgent string) error {
	// Check if domain is a valid URL
	u, err := url.Parse(domain)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return &invalidHostError{}
	}

	robot.domain = u

	// Sets the User Agent
	robot.userAgent = userAgent

	// Gets the content of the robots.txt file
	resp, err := http.Get(domain + "/robots.txt")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)

		robot.parse(bodyString)
	}

	return nil
}

// IsAllowed returns if the current path is visitable by the current User Agent
func (robot *Robots) IsAllowed(path string) bool {
	group := robot.getGroup(robot.userAgent)
	if group.rules == nil {
		group = robot.getGroup("*")
		if group.rules == nil {
			// Could not found limitation by user agent
			return true
		}
	}

	if strings.HasPrefix(path, "https://") {
		path = strings.Replace(path, "https://", "http://", 1)
	}

	domain := strings.Replace(robot.domain.String(), "https://", "http://", 1)

	if strings.HasPrefix(path, domain) {
		path = strings.Replace(path, domain, "", 1)
	}

	var result = true
	var resultPathLength = 0

	for _, rule := range group.rules {
		if rule.pattern != nil {
			// The first matching pattern takes precedence
			if rule.pattern.MatchString(path) {
				return rule.isAllowed
			}
		} else {
			// The longest matching path takes precedence
			if resultPathLength > len(rule.path) {
				continue
			}

			if strings.HasPrefix(path, rule.path) {
				result = rule.isAllowed
				resultPathLength = len(rule.path)
			}
		}
	}

	return result
}

func (robot *Robots) parse(content string) {
	currentUserAgent := "*"

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 1 {
			rule := strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])

			switch rule {
			case "user-agent":
				ua := strings.ToLower(val)
				if currentUserAgent != ua {
					currentUserAgent = ua
				}
				break
			case "allow":
				robot.addRule(currentUserAgent, val, true)
				break
			case "disallow":
				robot.addRule(currentUserAgent, val, false)
				break
			case "crawl-delay":
				robot.setCrawlDelay(currentUserAgent, val)
				break
			case "sitemap":
				robot.addSitemap(val)
				break
			}
		}
	}
}

func (robot *Robots) getGroup(userAgent string) *group {
	userAgent = strings.ToLower(userAgent)

	g, ok := robot.groups[userAgent]

	// Doesn't exist let's add it
	if !ok {
		g = &group{userAgent: userAgent}
		robot.groups[userAgent] = g
	}

	return g
}

func (robot *Robots) addRule(userAgent string, path string, isAllowed bool) error {
	group := robot.getGroup(userAgent)

	isPattern := isPattern(path)
	if isPattern {
		path = replaceSuffix(path, "%24", "%2524")
	}

	// Keep * escaped
	path = strings.Replace(path, "%2A", "%252A", -1)
	if unescapedPath, err := url.PathUnescape(path); err == nil {
		path = unescapedPath
	} else {
		path = strings.Replace(path, "%252A", "%2A", -1)
	}

	if isPattern {
		regexPattern, err := compilePattern(path)
		if err != nil {
			return err
		}

		group.rules = append(group.rules, &rule{
			isAllowed: isAllowed,
			path:      path,
			pattern:   regexPattern,
		})
	} else {
		group.rules = append(group.rules, &rule{
			isAllowed: isAllowed,
			path:      path,
			pattern:   nil,
		})
	}

	return nil
}

func (robot *Robots) setCrawlDelay(userAgent string, crawlDelay string) error {
	group := robot.getGroup(userAgent)

	delay, err := strconv.ParseFloat(crawlDelay, 64)
	if err != nil {
		return err
	}

	group.crawlDelay = time.Duration(delay * float64(time.Second))

	return nil
}

func (robot *Robots) getCrawlDelay() time.Duration {
	return robot.getCrawlDelayByUserAgent(robot.userAgent)
}

func (robot *Robots) getCrawlDelayByUserAgent(userAgent string) time.Duration {
	group := robot.getGroup(userAgent)
	return group.crawlDelay
}

func (robot *Robots) addSitemap(urlStr string) error {
	u, err := url.Parse(urlStr)
	// To validate the url
	if err != nil || u.Scheme == "" || u.Host == "" {
		return &invalidHostError{}
	}

	robot.sitemaps = append(robot.sitemaps, urlStr)

	return nil
}

func isPattern(path string) bool {
	return strings.IndexRune(path, '*') > -1 || strings.HasSuffix(path, "$")
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.Replace(pattern, "\\*", "(?:.*)", -1)

	pattern = replaceSuffix(pattern, "\\$", "$")
	pattern = replaceSuffix(pattern, "%24", "\\$")
	pattern = replaceSuffix(pattern, "%2524", "%24")

	pattern = strings.Replace(pattern, "%2A", "\\*", -1)

	return regexp.Compile(pattern)
}

func replaceSuffix(str, suffix, replacement string) string {
	if strings.HasSuffix(str, suffix) {
		return str[:len(str)-len(suffix)] + replacement
	}

	return str
}
