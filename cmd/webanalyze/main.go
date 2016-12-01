package main

import (
	"encoding/csv"
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
	update  bool
	useCSV  bool
	useJSON bool
	workers int
	apps    string
	host    string
	hosts   string
)

func init() {
	flag.BoolVar(&useCSV, "csv", false, "output as csv")
	flag.BoolVar(&useJSON, "json", false, "output as json")
	flag.BoolVar(&update, "update", false, "update apps file")
	flag.IntVar(&workers, "worker", 4, "number of worker")
	flag.StringVar(&apps, "apps", "apps.json", "app definition file.")
	flag.StringVar(&host, "host", "", "single host to test")
	flag.StringVar(&hosts, "hosts", "", "list of hosts to test, one host per line.")
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

	var (
		res       []webanalyze.Result
		out       *os.File
		outWriter *csv.Writer
	)

	if useCSV {
		out, err = os.Create("webanalyze-output.csv")
		if err != nil {
			log.Println("error creating file:", err)
			return
		}
		defer out.Close()

		outWriter = csv.NewWriter(out)
		defer outWriter.Flush()

		outWriter.Write([]string{"Host", "Category", "App"})
	}

	for result := range results {
		res = append(res, result)
		if !useJSON {
			log.Printf("[+] %v (%v):\n", result.Host, result.Duration)
			for _, a := range result.Matches {
				log.Printf("\t- %v\t - %v\n", a.AppName, a.App.Cats)
			}
			if len(result.Matches) <= 0 {
				log.Printf("\t<no results>\n")
			}
		}

		if useCSV {
			for _, m := range result.Matches {
				for _, c := range m.Cats {
					var catName string
					var ok bool
					if catName, ok = webanalyze.AppDefs.Cats[c]; !ok {
						catName = fmt.Sprintf("%d", c)
					}
					outWriter.Write(
						[]string{
							result.Host,
							catName,
							m.AppName,
						},
					)
				}
			}
			outWriter.Flush()
		}
	}

	if useJSON {
		b, _ := json.Marshal(res)
		log.Println(string(b))
	}
}
