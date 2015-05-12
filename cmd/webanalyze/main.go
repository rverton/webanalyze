package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"webanalyze"
)

var host = flag.String("host", "", "single host to test")
var hosts = flag.String("hosts", "hosts", "list of hosts to test, one host per line.")
var workers = flag.Int("worker", 4, "number of worker")
var update = flag.Bool("update", false, "update apps file")
var apps = flag.String("apps", "apps.json", "app definition file.")
var useJson = flag.Bool("json", false, "output as json")

func init() {
	if cpu := runtime.NumCPU(); cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}
}

func main() {
	var file io.ReadCloser
	var err error

	flag.Parse()

	if *update {
		err := webanalyze.DownloadFile(webanalyze.WAPPALYZER_URL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		log.Fatalln("app definition file updated from ", webanalyze.WAPPALYZER_URL)

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

	results, err := webanalyze.Init(*workers, file, *apps)

	if err != nil {
		log.Println("error initializing:", err)
	}

	log.Printf("Scanning with %v workers.", *workers)

	res := make(map[string]webanalyze.Result)

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
