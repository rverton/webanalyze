package main

import (
	"bufio"
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
	"sync"

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
	var (
		file io.ReadCloser
		err  error
		wa   *webanalyze.WebAnalyzer

		outWriter *csv.Writer
	)

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

	if wa, err = webanalyze.NewWebAnalyzer(apps); err != nil {
		log.Fatalf("initialization failed: %v", err)
	}

	if outputMethod == "csv" {
		outWriter = csv.NewWriter(os.Stdout)
		outWriter.Write([]string{"Host", "Category", "App", "Version"})

		defer outWriter.Flush()

	}

	printHeader()

	var wg sync.WaitGroup
	hosts := make(chan string)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {

			for host := range hosts {
				job := webanalyze.NewOnlineJob(host, "", nil, crawlCount, searchSubdomain)

				result := wa.Process(job)

				if result.Error != nil {
					fmt.Printf("%v: error: %v", result.Host, result.Error)
				}

				switch outputMethod {
				case "stdout":
					fmt.Printf("%v (%v):\n", result.Host, result.Duration)
					for _, a := range result.Matches {

						var categories []string

						for _, cid := range a.App.Cats {
							categories = append(categories, wa.CategoryById(cid))
						}

						fmt.Printf("    %v, %v (%v)\n", a.AppName, a.Version, strings.Join(categories, ", "))
					}
					if len(result.Matches) <= 0 {
						fmt.Printf("    <no results>\n")
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

			wg.Done()
		}()
	}

	// read hosts from file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hosts <- scanner.Text()
	}

	close(hosts)
	wg.Wait()
}

func printHeader() {
	printOption("webanalyze", "v"+webanalyze.VERSION)
	printOption("workers", workers)
	printOption("apps", apps)
	printOption("crawl count", crawlCount)
	printOption("search subdomains", searchSubdomain)
	fmt.Printf("\n")
}

func printOption(name string, value interface{}) {
	fmt.Fprintf(os.Stderr, " :: %-14s : %v\n", name, value)
}
