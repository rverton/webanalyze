package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/rverton/webanalyze"
)

var (
	update          bool
	outputMethod    string
	workers         int
	apps            string
	host            string
	hosts           string
	crawlCount      int
	searchSubdomain bool
)

func init() {
	flag.StringVar(&outputMethod, "output", "stdout", "output format (stdout|csv|json)")
	flag.BoolVar(&update, "update", false, "update apps file")
	flag.IntVar(&workers, "worker", 4, "number of worker")
	flag.StringVar(&apps, "apps", "apps.json", "app definition file.")
	flag.StringVar(&host, "host", "", "single host to test")
	flag.StringVar(&hosts, "hosts", "", "filename with hosts, one host per line.")
	flag.IntVar(&crawlCount, "crawl", 0, "links to follow from the root page (default 0)")
	flag.BoolVar(&searchSubdomain, "search", true, "searches all urls with same base domain (i.e. example.com and sub.example.com)")

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
		err = webanalyze.DownloadFile(webanalyze.WappalyzerURL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		log.Println("app definition file updated from ", webanalyze.WappalyzerURL)

		if host == "" && hosts == "" {
			return
		}

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

	results, err := webanalyze.Init(workers, file, apps, crawlCount, searchSubdomain)

	if err != nil {
		log.Fatal("error initializing:", err)
	}

	log.Printf("Scanning with %v workers.", workers)

	var (
		res       []webanalyze.Result
		outWriter *csv.Writer
	)

	if outputMethod == "csv" {
		outWriter = csv.NewWriter(os.Stdout)
		outWriter.Write([]string{"Host", "Category", "App", "Version"})

		defer outWriter.Flush()

	}

	for result := range results {
		res = append(res, result)

		if result.Error != nil {
			log.Printf("[-] Error for %v: %v", result.Host, result.Error)
		}

		switch outputMethod {
		case "stdout":
			log.Printf("[+] %v (%v):\n", result.Host, result.Duration)
			for _, a := range result.Matches {

				var categories []string

				for _, cid := range a.App.Cats {
					categories = append(categories, webanalyze.AppDefs.Cats[string(cid)].Name)
				}

				log.Printf("\t- %v, %v (%v)\n", a.AppName, a.Version, strings.Join(categories, ", "))
			}
			if len(result.Matches) <= 0 {
				log.Printf("\t<no results>\n")
			}

		case "csv":
			for _, m := range result.Matches {
				outWriter.Write(
					[]string{
						result.Host,
						strings.Join(m.CatNames, ","),
						m.AppName,
						m.Version,
					},
				)
			}
			outWriter.Flush()
		case "json":

			output := struct {
				Hostname string             `json:"hostname"`
				Matches  []webanalyze.Match `json:"matches"`
			}{
				result.Host,
				result.Matches,
			}

			b, err := json.Marshal(output)
			if err != nil {
				log.Printf("cannot marshal output: %v\n", err)
			}

			b = append(b, '\n')
			os.Stdout.Write(b)
		}
	}
}
