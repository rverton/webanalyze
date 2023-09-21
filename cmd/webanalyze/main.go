package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rverton/webanalyze"
)

var (
	update          bool
	outputMethod    string
	workers         int
	techsFilename   string
	host            string
	hosts           string
	crawlCount      int
	searchSubdomain bool
	silent          bool
	redirect        bool
)

func init() {
	flag.StringVar(&outputMethod, "output", "stdout", "output format (stdout|csv|json)")
	flag.BoolVar(&update, "update", false, "update technologies file to current dir")
	flag.IntVar(&workers, "worker", 4, "number of worker")
	flag.StringVar(&techsFilename, "apps", "technologies.json", "technologies definition file")
	flag.StringVar(&host, "host", "", "single host to test")
	flag.StringVar(&hosts, "hosts", "", "filename with hosts, one host per line.")
	flag.IntVar(&crawlCount, "crawl", 0, "links to follow from the root page (default 0)")
	flag.BoolVar(&searchSubdomain, "search", true, "searches all urls with same base domain (i.e. example.com and sub.example.com)")
	flag.BoolVar(&silent, "silent", false, "avoid printing header (default false)")
	flag.BoolVar(&redirect, "redirect", false, "follow http redirects (default false)")
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
		err = webanalyze.DownloadFile("technologies.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		if !silent {
			log.Println("app definition file updated")
		}

		if host == "" && hosts == "" {
			return
		}

	}

	// lookup technologies.json file
	techsFilename, err = lookupFolders(techsFilename)
	if err != nil {
		log.Fatalf("error: can not open apps file %s: %s", techsFilename, err)
	}

	// add header if output mode is csv
	if outputMethod == "csv" {
		outWriter = csv.NewWriter(os.Stdout)
		outWriter.Write([]string{"Host", "Category", "App", "Version"})

		defer outWriter.Flush()

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

	var wg sync.WaitGroup
	hosts := make(chan string)

	techsFile, err := os.Open(techsFilename)
	if err != nil {
		log.Fatalf("error: can not open apps file %s: %s", techsFilename, err)
	}
	defer techsFile.Close()

	if wa, err = webanalyze.NewWebAnalyzer(techsFile, nil); err != nil {
		log.Fatalf("initialization failed: %v", err)
	}

	if !silent {
		printHeader()
	}

	appsInfo, err := os.Stat(techsFilename)
	if err != nil {
		log.Fatalf("error: cant open %v: %v", techsFilename, err)
	}

	if appsInfo.ModTime().Before(time.Now().Add(24 * time.Hour * 7 * -1)) {
		log.Printf("warning: %v is older than a week", techsFilename)
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {

			for host := range hosts {
				job := webanalyze.NewOnlineJob(host, "", nil, crawlCount, searchSubdomain, redirect)
				result, links := wa.Process(job)

				if searchSubdomain {
					for _, v := range links {
						crawlJob := webanalyze.NewOnlineJob(v, "", nil, 0, false, redirect)
						result, _ := wa.Process(crawlJob)
						output(result, wa, outWriter)
					}
				}

				output(result, wa, outWriter)
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

func output(result webanalyze.Result, wa *webanalyze.WebAnalyzer, outWriter *csv.Writer) {
	if result.Error != nil {
		fmt.Fprintf(os.Stderr, "%v error: %v\n", result.Host, result.Error)
		return
	}

	switch outputMethod {
	case "stdout":
		fmt.Printf("%v (%.1fs):\n", result.Host, result.Duration.Seconds())
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

func printHeader() {
	printOption("webanalyze", "v"+webanalyze.VERSION)
	printOption("workers", workers)
	printOption("technologies", techsFilename)
	printOption("crawl count", crawlCount)
	printOption("search subdomains", searchSubdomain)
	printOption("follow redirects", redirect)
	fmt.Printf("\n")
}

func printOption(name string, value interface{}) {
	fmt.Fprintf(os.Stderr, " :: %-17s : %v\n", name, value)
}

func lookupFolders(filename string) (string, error) {
	if filepath.IsAbs(filename) {
		return filename, nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	executableDir := filepath.Dir(executable)

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	folders := []string{"./", executableDir, home}

	for _, folder := range folders {
		path := filepath.Join(folder, filename)

		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("could not find the technologies file: " + filename)
}
