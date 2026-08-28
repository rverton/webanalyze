package webanalyze

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestProcessRequestHeaders(t *testing.T) {
	var received http.Header
	var receivedHost string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		receivedHost = r.Host
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	analyzer, err := NewWebAnalyzer(strings.NewReader(`{"technologies":{},"categories":{}}`), server.Client())
	if err != nil {
		t.Fatalf("cannot create analyzer: %v", err)
	}

	headers := http.Header{
		"Accept":     []string{"text/html"},
		"Host":       []string{"tenant.example"},
		"User-Agent": []string{"custom-agent"},
		"X-Custom":   []string{"first", "second"},
	}
	job := NewOnlineJob(server.URL, "", nil, 0, false, false)
	job.RequestHeaders = headers
	result, _ := analyzer.Process(job)
	if result.Error != nil {
		t.Fatalf("Process returned an error: %v", result.Error)
	}

	if got, want := received.Get("Accept"), "text/html"; got != want {
		t.Errorf("got Accept header %q, want %q", got, want)
	}
	if got, want := received.Get("User-Agent"), "custom-agent"; got != want {
		t.Errorf("got User-Agent header %q, want %q", got, want)
	}
	if got, want := receivedHost, "tenant.example"; got != want {
		t.Errorf("got request host %q, want %q", got, want)
	}
	if got, want := received.Values("X-Custom"), []string{"first", "second"}; !slices.Equal(got, want) {
		t.Errorf("got X-Custom headers %q, want %q", got, want)
	}
	if headers.Get("Accept") != "text/html" {
		t.Error("Process modified the provided headers")
	}
}

func TestFetchHostDefaultAcceptHeader(t *testing.T) {
	var accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	resp, err := fetchHost(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("fetchHost returned an error: %v", err)
	}
	resp.Body.Close()

	if got, want := accept, "*/*"; got != want {
		t.Errorf("got Accept header %q, want %q", got, want)
	}
}

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
