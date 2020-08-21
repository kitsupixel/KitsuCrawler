package crawler

import (
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestGetDomain(t *testing.T) {
	testStr1, _ := url.Parse("http://www.kitsupixel.pt")
	testStr2, _ := url.Parse("http://pplware.sapo.pt")
	testStr3, _ := url.Parse("http://www.example.co.uk")
	testStr4, _ := url.Parse("http://example.com")

	domain, domainName := getDomain(testStr1)
	if domain != "kitsupixel.pt" {
		t.Fatalf("%s != %s", domain, "kitsupixel.pt")
	}

	if domainName != "kitsupixel" {
		t.Fatalf("%s != %s", domainName, "kitsupixel")
	}

	domain, domainName = getDomain(testStr2)
	if domain != "pplware.sapo.pt" {
		t.Fatalf("%s != %s", domain, "pplware.sapo.pt")
	}

	if domainName != "pplware.sapo" {
		t.Fatalf("%s != %s", domainName, "pplware.sapo")
	}

	domain, domainName = getDomain(testStr3)
	if domain != "example.co.uk" {
		t.Fatalf("%s != %s", domain, "example.co.uk")
	}

	if domainName != "example" {
		t.Fatalf("%s != %s", domainName, "example")
	}

	domain, domainName = getDomain(testStr4)
	if domain != "example.com" {
		t.Fatalf("%s != %s", domain, "example.com")
	}

	if domainName != "example" {
		t.Fatalf("%s != %s", domainName, "example")
	}
}

func TestRemoveDuplicatesFromFile(t *testing.T) {
	testFilePath := "./../outputs/test.txt"

	f, _ := os.Create(testFilePath)

	_, _ = f.WriteString("test1" + "\n")
	_, _ = f.WriteString("test2" + "\n")
	_, _ = f.WriteString("test1" + "\n")
	_, _ = f.WriteString("test2" + "\n")
	_, _ = f.WriteString("test3" + "\n")

	_ = f.Close()

	removeDuplicatesFromFile(testFilePath)

	line, _ := ioutil.ReadFile(testFilePath)
	// turn the byte slice into string format
	strLine := string(line)
	// split the lines by a space, can also change this
	lines := strings.Split(strLine, "\n")

	lines = deleteEmpty(lines)

	if len(lines) > 3 {
		t.Fatalf("%d != 3 ", len(lines))
	}

	if lines[0] != "test1" {
		t.Fatalf("%s != %s", lines[0], "test1")
	}

	if lines[1] != "test2" {
		t.Fatalf("%s != %s", lines[1], "test2")
	}

	if lines[2] != "test3" {
		t.Fatalf("%s != %s", lines[2], "test3")
	}

	os.Remove(testFilePath)
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
