package webanalyze

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	// AppDefs provides access to the unmarshalled apps.json file
	AppDefs *AppsDefinition
	timeout = 8 * time.Second
)

// Result type encapsulates the result information from a given host
type Result struct {
	Host     string        `json:"host"`
	Matches  []Match       `json:"matches"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error"`
}

// Match type encapsulates the App information from a match on a document
type Match struct {
	App     `json:"app"`
	AppName string     `json:"app_name"`
	Matches [][]string `json:"matches"`
	Version string     `json:"version"`
}

// WebAnalyzer types holds an analyzation job
type WebAnalyzer struct {
	Results chan Result
	jobs    chan *Job
	wg      *sync.WaitGroup
}

func (m *Match) updateVersion(version string) {
	if version != "" {
		m.Version = version
	}
}

// Init sets up all the workders, reads in the host data and returns the results channel or an error
func Init(workers int, hosts io.Reader, appsFile string) (chan Result, error) {
	wa, err := NewWebAnalyzer(workers, appsFile)
	if err != nil {
		return nil, err
	}
	// send hosts line by line to worker channel
	go func(hosts io.Reader, wa *WebAnalyzer) {
		scanner := bufio.NewScanner(hosts)
		for scanner.Scan() {
			url := scanner.Text()
			wa.schedule(NewOnlineJob(url, "", nil))
		}
		// wait for workers to finish, the close result channel to signal finish of scan
		wa.close()
	}(hosts, wa)
	return wa.Results, nil
}

// NewWebAnalyzer returns an analyzer struct for an ongoing job, which may be
// "fed" jobs via a method and returns them via a channel when complete.
func NewWebAnalyzer(workers int, appsFile string) (*WebAnalyzer, error) {
	wa := new(WebAnalyzer)
	wa.Results = make(chan Result)
	wa.jobs = make(chan *Job)
	wa.wg = new(sync.WaitGroup)
	if err := loadApps(appsFile); err != nil {
		return nil, err
	}
	// start workers
	initWorker(workers, wa.jobs, wa.Results, wa.wg)
	return wa, nil
}

func (wa *WebAnalyzer) schedule(job *Job) {
	wa.jobs <- job
}

func (wa *WebAnalyzer) close() {
	close(wa.jobs)
	wa.wg.Wait()
	close(wa.Results)
}

// start n worker and let them listen on channel c for hosts to scan
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

func fetchHost(host string) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}

	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// do http request and analyze response
func process(job *Job) ([]Match, error) {
	var apps = make([]Match, 0)
	var err error

	var cookies []*http.Cookie
	var cookiesMap = make(map[string]string)
	var body []byte
	var headers http.Header

	// get response from host if allowed
	if job.forceNotDownload {
		body = job.Body
		headers = job.Headers
		cookies = job.Cookies
	} else {
		resp, err := fetchHost(job.URL)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve")
		}

		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			headers = resp.Header
			cookies = resp.Cookies()
		}
	}

	for _, c := range cookies {
		cookiesMap[c.Name] = c.Value
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
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
		if m, v := findMatches(string(body), app.HTMLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
			findings.updateVersion(v)
		}

		// check response header
		headerFindings, version := app.FindInHeaders(headers)
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

		// check cookies
		for _, c := range app.CookieRegex {
			if _, ok := cookiesMap[c.Name]; ok {

				// if there is a regexp set, ensure it matches.
				// otherwise just add this as a match
				if c.Regexp != nil {

					// only match single AppRegexp on this specific cookie
					if m, v := findMatches(cookiesMap[c.Name], []AppRegexp{c}); len(m) > 0 {
						findings.Matches = append(findings.Matches, m...)
						findings.updateVersion(v)
					}

				} else {
					findings.Matches = append(findings.Matches, []string{c.Name})
				}
			}

		}

		if len(findings.Matches) > 0 {
			apps = append(apps, findings)

			// handle implies
			for _, implies := range app.Implies {
				for implyAppname, implyApp := range AppDefs.Apps {
					if implies != implyAppname {
						continue
					}

					f2 := Match{
						App:     implyApp,
						AppName: implyAppname,
						Matches: make([][]string, 0),
					}
					apps = append(apps, f2)
				}

			}
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
