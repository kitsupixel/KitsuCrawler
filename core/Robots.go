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

type Rule struct {
	isAllowed bool
	path      string
	pattern   *regexp.Regexp
}

type Group struct {
	userAgent  string
	rules      []*Rule
	crawlDelay time.Duration
}

type Robots struct {
	domain    *url.URL
	groups    map[string]*Group
	sitemaps  []string
	host      string
	userAgent string
}

type InvalidHostError struct{}

func (e InvalidHostError) Error() string {
	return "URL is not valid for this robots.txt file"
}

func NewRobot() *Robots {
	return &Robots{
		groups: make(map[string]*Group),
	}
}

func (robot *Robots) Build(domain string, userAgent string) error {
	// Check if domain is a valid URL
	u, err := url.Parse(domain)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return &InvalidHostError{}
	}

	robot.domain = u

	// Sets the User Agent
	robot.userAgent = userAgent

	// Gets the content of the robots.txt file
	resp, _ := http.Get(domain + "/robots.txt")
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

func (robot *Robots) IsAllowed(path string) bool {
	group := robot.getGroup(robot.userAgent)
	if group.rules == nil {
		group = robot.getGroup("*")
		if group.rules == nil {
			// Could not found limitation by user agent
			return true
		}
	}

	if strings.HasPrefix(path, robot.domain.String()) {
		path = strings.Replace(path, robot.domain.String(), "", 1)
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

func (robot *Robots) getGroup(userAgent string) *Group {
	userAgent = strings.ToLower(userAgent)

	group, ok := robot.groups[userAgent]

	// Doesn't exist let's add it
	if !ok {
		group = &Group{userAgent: userAgent}
		robot.groups[userAgent] = group
	}

	return group
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

		group.rules = append(group.rules, &Rule{
			isAllowed: isAllowed,
			path:      path,
			pattern:   regexPattern,
		})
	} else {
		group.rules = append(group.rules, &Rule{
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

func (robot *Robots) getCrawlDelay(userAgent string) time.Duration {
	group := robot.getGroup(userAgent)
	return group.crawlDelay
}

func (robot *Robots) addSitemap(urlStr string) error {
	u, err := url.Parse(urlStr)
	// To validate the url
	if err != nil || u.Scheme == "" || u.Host == "" {
		return &InvalidHostError{}
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


