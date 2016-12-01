package webanalyze

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// start n worker and let them listen on c for hosts to scan
func initWorker(count int, c chan string, results chan Result) {
	// start workers based on flag
	for i := 0; i < count; i++ {
		wg.Add(1)
		go worker(i, c, results)
	}
}

// worker loops until channel is closed. processes a single host at once
func worker(i int, c chan string, results chan Result) {

	for host := range c {

		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = fmt.Sprintf("http://%s", host)
		}

		t0 := time.Now()
		result, err := process(host)
		t1 := time.Now()

		res := Result{
			Host:     host,
			Matches:  result,
			Duration: t1.Sub(t0),
			Error:    err,
		}

		results <- res

	}

	wg.Done()
}

// do http request and analyze response
func process(host string) ([]Match, error) {
	var apps = make([]Match, 0)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(host)
	if err != nil {
		return apps, err
	}
	defer resp.Body.Close()

	// ignore error, body/document not always needed
	body, _ := ioutil.ReadAll(resp.Body)
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(body))

	for appname, app := range AppDefs.Apps {

		findings := Match{
			AppName:    appname,
			AppWebsite: app.Website,
			Matches:    make([][]string, 0),
		}

		// check raw html
		if m := findMatches(string(body), app.HTMLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
		}

		// check response header
		for _, h := range app.HeaderRegex {
			headerValue := resp.Header.Get(h.Name)

			if headerValue == "" {
				continue
			}

			if m := findMatches(headerValue, []*regexp.Regexp{h.Regex}); len(m) > 0 {
				findings.Matches = append(findings.Matches, m...)
			}

		}

		// check url
		if m := findMatches(resp.Request.URL.String(), app.URLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
		}

		// check script tags
		doc.Find("script").Each(func(i int, s *goquery.Selection) {

			if script, exists := s.Attr("src"); exists {
				if m := findMatches(script, app.ScriptRegex); len(m) > 0 {
					findings.Matches = append(findings.Matches, m...)
				}
			}

		})

		// check meta tags
		for _, h := range app.MetaRegex {

			selector := fmt.Sprintf("meta[name='%s']", h.Name)

			doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				content, _ := s.Attr("content")

				if m := findMatches(content, []*regexp.Regexp{h.Regex}); len(m) > 0 {
					findings.Matches = append(findings.Matches, m...)
				}

			})

		}

		if len(findings.Matches) > 0 {
			apps = append(apps, findings)
		}

	}

	return apps, nil
}

// runs a list of regexes on content
func findMatches(content string, regexes []*regexp.Regexp) [][]string {
	var m [][]string
	for _, r := range regexes {
		matches := r.FindAllStringSubmatch(content, -1)
		if matches != nil {
			m = append(m, matches...)
		}

	}

	return m
}
