package core

import "net/url"

type Crawler struct {
	domain *url.URL
	depthLevel int32
	ignoreRobots bool
}