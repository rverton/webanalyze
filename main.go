package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup
var appDefs *AppsDefinition

var host = flag.String("host", "", "single host to test")
var hosts = flag.String("hosts", "hosts", "list of hosts to test, one host per line.")
var workers = flag.Int("worker", 4, "number of worker")
var update = flag.Bool("update", false, "update apps file")
var apps = flag.String("apps", "apps.json", "app definition file.")
var useJson = flag.Bool("json", false, "output as json")

type Result struct {
	Host     string        `json:"-"`
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

	err = loadApps(*apps)
	if err != nil {
		log.Fatalf("error: can not load app definition file: %v", err)
	}

	log.Printf("Loaded %v app definitions", len(appDefs.Apps))
	log.Printf("Scanning with %v workers.", *workers)

	// start worker
	initWorker(*workers, c, results)

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

	res := make(map[string]Result)

	for result := range results {
		if !*useJson {
			fmt.Printf("[+] %v (%v):\n", result.Host, result.Duration)
			for _, a := range result.Matches {
				fmt.Printf("\t- %v\n", a.AppName)
			}
			if len(result.Matches) <= 0 {
				fmt.Printf("\t<no results>\n")
			}
		} else {
			res[result.Host] = result
		}
	}

	if *useJson {
		b, _ := json.Marshal(res)
		fmt.Println(string(b))
	}

}
