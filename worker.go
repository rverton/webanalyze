package webanalyze

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var timeout = time.Duration(5 * time.Second)

// start n worker and let them listen on c for hosts to scan
func initWorker(count int, c chan *Job, results chan Result, wg *sync.WaitGroup) {
	// start workers based on flag
	for i := 0; i < count; i++ {
		wg.Add(1)
		go worker(c, results, wg)
	}
}

// worker loops until channel is closed. processes a single host at once
func worker(c chan *Job, results chan Result, wg *sync.WaitGroup) {
	for job := range c {
		if !strings.HasPrefix(job.URL, "http://") && !strings.HasPrefix(job.URL, "https://") {
			job.URL = fmt.Sprintf("http://%s", job.URL)
		}

		t0 := time.Now()
		result, err := process(job)
		t1 := time.Now()

		res := Result{
			Host:     job.URL,
			Matches:  result,
			Duration: t1.Sub(t0),
			Error:    err,
		}
		results <- res
	}
	wg.Done()
}

func fetchHost(host string) ([]byte, *http.Header, error) {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	resp, err := client.Get(host)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// ignore error, body/document not always needed
		return nil, &resp.Header, nil
	}
	return body, &resp.Header, nil
}

// do http request and analyze response
func process(job *Job) ([]Match, error) {
	var apps = make([]Match, 0)

	if (job.Body == nil || len(job.Body) == 0) && !job.forceNotDownload {
		_body, headers, err := fetchHost(job.URL)
		if err != nil {
			return nil, err
		}
		job.Body = _body
		job.Headers = *headers
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(job.Body))
	if err != nil {
		return nil, err
	}

	for appname, app := range AppDefs.Apps {
		// TODO: Reduce complexity in this for-loop by functionalising out
		// the sub-loops and checks.

		findings := Match{
			App:     app,
			AppName: appname,
			Matches: make([][]string, 0),
		}

		// check raw html
		if m, v := findMatches(string(job.Body), app.HTMLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
			findings.updateVersion(v)
		}

		// check response header
		headerFindings, version := app.FindInHeaders(job.Headers)
		findings.Matches = append(findings.Matches, headerFindings...)
		findings.updateVersion(version)

		// check url
		if m, v := findMatches(job.URL, app.URLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
			findings.updateVersion(v)
		}

		// check script tags
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			if script, exists := s.Attr("src"); exists {
				if m, v := findMatches(script, app.ScriptRegex); len(m) > 0 {
					findings.Matches = append(findings.Matches, m...)
					findings.updateVersion(v)
				}
			}
		})

		// check meta tags
		for _, h := range app.MetaRegex {
			selector := fmt.Sprintf("meta[name='%s']", h.Name)
			doc.Find(selector).Each(func(i int, s *goquery.Selection) {
				content, _ := s.Attr("content")
				if m, v := findMatches(content, []AppRegexp{h}); len(m) > 0 {
					findings.Matches = append(findings.Matches, m...)
					findings.updateVersion(v)
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
func findMatches(content string, regexes []AppRegexp) ([][]string, string) {
	var m [][]string
	var version string

	for _, r := range regexes {
		matches := r.Regexp.FindAllStringSubmatch(content, -1)
		if matches == nil {
			continue
		}

		m = append(m, matches...)

		if r.Version != "" {
			version = findVersion(m, r.Version)
		}

	}
	return m, version
}

// parses a version against matches
func findVersion(matches [][]string, version string) string {
	/*
		log.Printf("Matches: %v", matches)
		log.Printf("Version: %v", version)
	*/

	var v string

	for _, matchPair := range matches {
		// replace backtraces (max: 3)
		for i := 1; i <= 3; i++ {
			bt := fmt.Sprintf("\\%v", i)
			if strings.Contains(version, bt) && len(matchPair) >= i {
				v = strings.Replace(version, bt, matchPair[i], 1)
			}
		}

		// return first found version
		if v != "" {
			return v
		}

	}

	return ""
}
