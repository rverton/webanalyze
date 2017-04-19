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
	update   bool
	csvFile  string
	jsonFile string
	workers  int
	apps     string
	host     string
	hosts    string
)

func init() {
	flag.StringVar(&csvFile, "csv", "", "export to csv file")
	flag.StringVar(&jsonFile, "json", "", "output to json file")
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
		err = webanalyze.DownloadFile(webanalyze.WappalyzerURL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		log.Fatalln("app definition file updated from ", webanalyze.WappalyzerURL)

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
		res        []webanalyze.Result
		out        *os.File
		outWriter  *csv.Writer
		outputMode string
	)

	if csvFile != "" {
		outputMode = "csv"
		out, err = os.Create(csvFile)
		outWriter = csv.NewWriter(out)
		outWriter.Write([]string{"Host", "Category", "App"})

		defer outWriter.Flush()

	} else if jsonFile != "" {
		outputMode = "json"
		out, err = os.Create(jsonFile)
		out.Write([]byte("["))

		defer func() {
			out.Seek(-1, os.SEEK_END)
			out.Write([]byte("]"))
		}()
	} else {
		outputMode = "stdout"
	}

	if err != nil {
		log.Println("error creating export file:", err)
		return
	}

	for result := range results {
		res = append(res, result)

		switch outputMode {
		case "stdout":
			log.Printf("[+] %v (%v):\n", result.Host, result.Duration)
			for _, a := range result.Matches {

				var categories []string

				for _, cid := range a.App.Cats {
					categories = append(categories, webanalyze.AppDefs.Cats[cid].Name)
				}

				log.Printf("\t- %v (%v)\n", a.AppName, strings.Join(categories, ", "))
			}
			if len(result.Matches) <= 0 {
				log.Printf("\t<no results>\n")
			}

		case "csv":
			for _, m := range result.Matches {
				for _, c := range m.Cats {
					var catName string
					if category, ok := webanalyze.AppDefs.Cats[c]; !ok {
						catName = fmt.Sprintf("%d", category.Name)
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
		case "json":
			b, err := json.Marshal(res)
			if err != nil {
				log.Printf("error marshaling content: %v\n", err)
			}

			out.Write(b)
			out.Write([]byte(","))
		}
	}
}
