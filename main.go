package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup
var appDefs *AppsDefinition

var host = flag.String("host", "", "single host to test")
var hosts = flag.String("hosts", "hosts", "list of hosts, one line per host.")
var workers = flag.Int("worker", 50, "number of worker")
var update = flag.Bool("update", false, "update apps file")
var apps = flag.String("apps", "apps.json", "app definition file.")
var useJson = flag.Bool("json", false, "output as json")

type Result struct {
	Host     string        `json:"host"`
	Matches  []Match       `json:"matches"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error"`
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var file io.ReadCloser
	var err error

	results := make(chan Result)
	c := make(chan string)

	flag.Parse()

	if *update {
		err := downloadFile(WAPPALYZER_URL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		log.Fatalln("app definition file updated from ", WAPPALYZER_URL)

	}

	// check single host or hosts file
	if *host != "" {
		file = ioutil.NopCloser(strings.NewReader(*host))
	} else {
		file, err = os.Open(*hosts)
		if err != nil {
			log.Fatalf("error: can not open host file %s: %s", *hosts, err)
		}
	}
	defer file.Close()

	err = LoadApps(*apps)
	if err != nil {
		log.Fatalf("error: can not load app definition file: %v", err)
	}

	log.Printf("Loaded %v app definitions", len(appDefs.Apps))
	log.Printf("Scanning with %v workers.", *workers)

	// start worker
	InitWorker(*workers, c, results)

	// send hosts line by line to worker channel
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			c <- scanner.Text()
		}
		close(c)

		// wait for workers to finish, the close result channel to signal finish of scan
		wg.Wait()
		close(results)
	}()

	var res []Result

	for result := range results {
		if !*useJson {
			fmt.Printf("[+] %v (%v):\n", result.Host, result.Duration)
			for _, a := range result.Matches {
				fmt.Printf("\t- %v\n", a.AppName)
			}
		} else {
			res = append(res, result)
		}
	}

	if *useJson {
		b, _ := json.Marshal(res)
		fmt.Println(string(b))
	}

}

// start n worker and let them listen on c for hosts to scan
func InitWorker(count int, c chan string, results chan Result) {
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

	body, err := ioutil.ReadAll(resp.Body)

	for appname, app := range appDefs.Apps {

		findings := Match{
			AppName:    appname,
			AppWebsite: app.Website,
			Matches:    make([][]string, 0),
		}

		// Test HTML
		if m := findMatches(string(body), app.HTMLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
		}

		// Test Header
		for headerName, r := range app.HeaderRegex {
			if m := findMatches(resp.Header.Get(headerName), []*regexp.Regexp{r}); len(m) > 0 {
				findings.Matches = append(findings.Matches, m...)
			}

		}

		// Test URL
		if m := findMatches(resp.Request.URL.String(), app.URLRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
		}

		// Test Scripts
		if m := findMatches(string(body), app.ScriptRegex); len(m) > 0 {
			findings.Matches = append(findings.Matches, m...)
		}

		if len(findings.Matches) > 0 {
			apps = append(apps, findings)
		}

	}

	return apps, nil
}

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
