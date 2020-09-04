package webanalyze

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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

	u, _ := url.Parse("http://127.0.0.1")

	links := parseLinks(doc, u, false)
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

func TestParseLinksSubdomain(t *testing.T) {

	crawlData := `
	<html><body>
	<a href="https://example.com">google.com</a>
	<a href="https://foo.example.com">robinverton.de</a>
	<a href="https://bar.foo.example.com">robinverton.de</a>
	<a href="http://127.0.0.1/foobar.html">Same Host</a>
	</body></html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(crawlData))
	if err != nil {
		t.Fatalf("Invalid testing document")
	}

	u, _ := url.Parse("http://example.com")

	if links := parseLinks(doc, u, false); len(links) != 0 {
		fmt.Println(links)
		t.Fatalf("Invalid number of subdomain crawl returned")
	}

	if len(parseLinks(doc, u, true)) != 2 {
		t.Fatalf("Invalid number of subdomain crawl returned")
	}

	return
}

func TestIsSubdomain(t *testing.T) {
	u1, _ := url.Parse("http://example.com")

	u2, _ := url.Parse("http://sub.example.com")

	u3, _ := url.Parse("http://sub1.sub2.example.com")

	if !isSubdomain(u1, u2) {
		t.Fatalf("%v is not a subdomain of %v (but should be)", u2, u1)
	}

	if !isSubdomain(u1, u3) {
		t.Fatalf("%v is not a subdomain of %v (but should be)", u2, u1)
	}
}

func TestRedirect(t *testing.T) {
	testServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not be reached"))
	}))

	testServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, testServer1.URL, http.StatusTemporaryRedirect)
	}))

	defer func() {
		testServer1.Close()
		testServer2.Close()
	}()

	resp, err := fetchHost(testServer2.URL, nil)

	if err != nil {
		t.Fatal(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "should not be reached" {
		t.Error("fetchHost did follow redirect")
	}

}
