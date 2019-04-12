package webanalyze

import (
	"testing"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

func TestParseLinks(t *testing.T) {

	crawlData := `
	<html><body>
	<a href="./foo.html">Relative Link 1</a>
	<a href="https://google.com">google.com</a>
	<a href="https://robinverton.de">robinverton.de</a>
	<a href="http://127.0.0.1/foobar.html">Same Host</a>
	</body></html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(crawlData))
	if err != nil {
		t.Fatalf("Invalid testing document")
	}

	links := parseLinks(doc, "http://127.0.0.1")
	if len(links) != 2 {
		t.Fatalf("Invalid number of links returned")
	}

	if links[0] != "http://127.0.0.1/foo.html" {
		t.Fatalf("Invalid link parsed")
	}

	if links[1] != "http://127.0.0.1/foobar.html" {
		t.Fatalf("Invalid link parsed")
	}

	return
}