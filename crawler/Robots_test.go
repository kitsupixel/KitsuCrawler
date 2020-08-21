package crawler

import (
	"net/url"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	robot := NewRobot()

	robot.Build("https://www.kitsupixel.pt", "TestBot")

	if !robot.IsAllowed("https://www.kitsupixel.pt/4") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/4")
	}

	if robot.IsAllowed("https://www.kitsupixel.pt/admin/login") {
		t.Fatalf("%s shouldn't be allowed in here", "https://www.kitsupixel.pt/admin/login")
	}

	if !robot.IsAllowed("https://www.kitsupixel.pt/4") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/4")
	}

	robot.userAgent = "Googlebot"

	if !robot.IsAllowed("https://www.kitsupixel.pt/admin/login") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/admin/login")
	}

	if robot.IsAllowed("https://www.kitsupixel.pt/nogooglebot/index.html") {
		t.Fatalf("%s shouldn't be allowed in here", "https://www.kitsupixel.pt/nogooglebot/index.html")
	}

	if robot.IsAllowed("https://www.kitsupixel.pt/test.gif") {
		t.Fatalf("%s shouldn't be allowed in here", "https://www.kitsupixel.pt/test.gif")
	}
}

func TestBebitusBuild(t *testing.T) {
	robot := NewRobot()

	robot.Build("https://www.bebitus.pt", "TestBot")

	if !robot.IsAllowed("https://www.bebitus.pt/un-conte-de-fil") {
		t.Fatalf("%s should be allowed in here", "https://www.bebitus.pt/un-conte-de-fil")
	}

	if robot.IsAllowed("https://www.bebitus.pt/search/?rs=true&q=herzlisten:%22PromoCole%22&&icn=BEB_PT_20200817_PTbacktoschool_index&ici=onpage_banner_brandcorner_backtoschoolpromo") {
		t.Fatalf("%s shouldn't be allowed in here", "http://www.bebitus.pt/search/?rs=true&q=herzlisten:%22PromoCole%22&&icn=BEB_PT_20200817_PTbacktoschool_index&ici=onpage_banner_brandcorner_backtoschoolpromo")
	}

	if robot.IsAllowed("http://www.bebitus.pt/search/?rs=true&q=herzlisten:%22PromoCole%22&&icn=BEB_PT_20200817_PTbacktoschool_index&ici=onpage_banner_brandcorner_backtoschoolpromo") {
		t.Fatalf("%s shouldn't be allowed in here", "http://www.bebitus.pt/search/?rs=true&q=herzlisten:%22PromoCole%22&&icn=BEB_PT_20200817_PTbacktoschool_index&ici=onpage_banner_brandcorner_backtoschoolpromo")
	}
}

func TestNoRobots(t *testing.T) {
	robot := NewRobot()

	robot.Build("http://www.copiabite.pt", "TestBot")

	if !robot.IsAllowed("http://www.copiabite.pt/4") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/4")
	}

	if !robot.IsAllowed("http://www.copiabite.pt/admin/login") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/admin/login")
	}

	if !robot.IsAllowed("http://www.copiabite.pt/4") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/4")
	}

	robot.userAgent = "Googlebot"

	if !robot.IsAllowed("http://www.copiabite.pt/admin/login") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/admin/login")
	}

	if !robot.IsAllowed("http://www.copiabite.pt/nogooglebot/index.html") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/nogooglebot/index.html")
	}

	if !robot.IsAllowed("http://www.copiabite.pt/test.gif") {
		t.Fatalf("%s should be allowed in here", "https://www.kitsupixel.pt/test.gif")
	}
}

func TestBuildFail(t *testing.T) {
	robot := NewRobot()

	err := robot.Build("Testing Fail", "TestBot")
	if err == nil {
		t.Fatalf("It should have failed!")
	}
}

func TestIsAllowed(t *testing.T) {
	robot := NewRobot()

	robot.domain, _ = url.Parse("http://www.example.com")
	robot.userAgent = "TestBot"

	robot.addRule("TestBot", "/test", false)
	robot.addRule("TestBot", "/", true)

	if robot.IsAllowed("http://www.example.com/test/test.html") {
		t.Fatalf("%s shouldn't be allowed in here", "http://www.example.com/test/test.html")
	}

	if robot.IsAllowed("/test/test.html") {
		t.Fatalf("%s shouldn't be allowed in here", "http://www.example.com/test/test.html")
	}

	if !robot.IsAllowed("http://www.example.com/contacts.html") {
		t.Fatalf("%s should be allowed in here", "http://www.example.com/contacts.html")
	}

	if !robot.IsAllowed("/contacts.html") {
		t.Fatalf("%s should be allowed in here", "http://www.example.com/contacts.html")
	}
}

func TestAddRule(t *testing.T) {
	robot := NewRobot()

	robot.addRule("TestBot", "/", true)
	robot.addRule("TestBot", "/test", false)
	robot.addRule("TestBot", "*", false)

	rules := robot.getGroup("TestBot").rules

	if !rules[0].isAllowed {
		t.Fatalf("Rule 1 is meant to be true")
	}

	if rules[1].isAllowed {
		t.Fatalf("Rule 2 is meant to be false")
	}

	if rules[2].isAllowed {
		t.Fatalf("Rule 3 is meant to be false")
	}

	if rules[0].path != "/" {
		t.Fatalf("%s != %s", rules[0].path, "/")
	}

	if rules[1].path != "/test" {
		t.Fatalf("%s != %s", rules[1].path, "/test")
	}

	if rules[2].pattern == nil {
		t.Fatalf("%s is not meant to be nill", rules[1].path)
	}
}

func TestSetCrawlerDelay(t *testing.T) {
	robot := NewRobot()

	robot.setCrawlDelay("TestBot", "10")

	delay := robot.getCrawlDelayByUserAgent("TestBot")

	if delay != time.Duration(10*float64(time.Second)) {
		t.Fatalf("%s != %s", delay, time.Duration(10*float64(time.Second)))
	}
}

func TestAddSitemap(t *testing.T) {
	robot := NewRobot()

	robot.addSitemap("http://www.example.com/sitemap1.xml")

	if robot.sitemaps[0] != "http://www.example.com/sitemap1.xml" {
		t.Fatalf("%s != %s", robot.sitemaps[0], "http://www.example.com/sitemap1.xml")
	}

	robot.addSitemap("http://www.example.com/sitemap2.xml")

	if robot.sitemaps[1] != "http://www.example.com/sitemap2.xml" {
		t.Fatalf("%s != %s", robot.sitemaps[1], "http://www.example.com/sitemap2.xml")
	}
}

func TestFailAddSitemap(t *testing.T) {
	robot := NewRobot()

	err := robot.addSitemap("/Pokemon")

	if err == nil {
		t.Fatalf("It should fail")
	}
}
