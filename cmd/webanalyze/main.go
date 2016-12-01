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

	"github.com/rverton/webanalyze"
)

var (
	workers int
	host    string
	hosts   string
	apps    string
	update  bool
	useJSON bool
)

func init() {
	flag.StringVar(&host, "host", "", "single host to test")
	flag.StringVar(&hosts, "hosts", "hosts", "list of hosts to test, one host per line.")
	flag.IntVar(&workers, "worker", 4, "number of worker")
	flag.BoolVar(&update, "update", false, "update apps file")
	flag.StringVar(&apps, "apps", "apps.json", "app definition file.")
	flag.BoolVar(&useJSON, "json", false, "output as json")
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

	if !update && host == "" && hosts == "" {
		flag.Usage()
		return
	}

	if update {
		err = webanalyze.DownloadFile(webanalyze.WAPPALYZER_URL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		log.Fatalln("app definition file updated from ", webanalyze.WAPPALYZER_URL)

	}

	// check single host or hosts file
	if host != "" {
		file = ioutil.NopCloser(strings.NewReader(host))
	} else {
		file, err = os.Open(hosts)
		if err != nil {
			log.Fatalf("error: can not open host file %s: %s", hosts, err)
		}
	}
	defer file.Close()

	results, err := webanalyze.Init(workers, file, apps)

	if err != nil {
		log.Println("error initializing:", err)
	}

	log.Printf("Scanning with %v workers.", workers)

	var res []webanalyze.Result

	for result := range results {
			fmt.Printf("[+] %v (%v):\n", result.Host, result.Duration)
		if !useJSON {
			for _, a := range result.Matches {
				fmt.Printf("\t- %v\n", a.AppName)
			}
			if len(result.Matches) <= 0 {
				fmt.Printf("\t<no results>\n")
			}
		} else {
			res = append(res, result)
		}
	}

	if useJSON {
		b, _ := json.Marshal(res)
		log.Println(string(b))
	}
}
