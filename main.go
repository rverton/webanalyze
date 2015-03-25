package main

import (
	"bufio"
	"crypto/tls"
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

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var file io.ReadCloser
	var err error

	c := make(chan string)

	var host = flag.String("host", "", "Single host to test")
	var hosts = flag.String("hosts", "hosts", "List of hosts. One line per host.")
	var workers = flag.Int("worker", 50, "Number of worker.")
	var update = flag.Bool("update", false, "Update apps file")
	var apps = flag.String("apps", "apps.json", "app definition file.")

	flag.Parse()

	if *update {
		err := downloadFile(WAPPALYZER_URL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}
		log.Println("app definition file updated from %v", WAPPALYZER_URL)
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

	appDefs, err = loadApps(*apps)
	if err != nil {
		log.Fatalf("error: can not load app definition file: %v", err)
	}

	log.Printf("Loaded %v app definitions", len(appDefs.Apps))
	log.Printf("Scanning with %v workers.", *workers)

	// start workers based on flag
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(i, c)
	}

	// send hosts line by line to worker channel
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c <- scanner.Text()
	}

	close(c)
	wg.Wait()
}

// worker loops until channel is closed and processes a single host
func worker(i int, c chan string) {

	for host := range c {

		url := host

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = fmt.Sprintf("http://%s", url)
		}

		t0 := time.Now()
		result, err := process(url)
		t1 := time.Now()
		diff := t1.Sub(t0)

		if err != nil {
			fmt.Printf("[-] %v: %v (%v, worker %v)\n", host, err, diff, i)
		} else {
			fmt.Printf("[+] %v (%v, worker %v):\n", host, diff, i)
			for _, a := range result {
				fmt.Printf("\t- %v\n", a.AppName)
			}
		}

	}

	wg.Done()
}

func process(host string) ([]Match, error) {
	var apps []Match

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
