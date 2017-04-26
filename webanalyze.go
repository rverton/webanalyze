package webanalyze

import (
	"bufio"
	"io"
	"sync"
	"time"
)

var (
	// AppDefs provides access to the unmarshalled apps.json file
	AppDefs *AppsDefinition
)

// Result type encapsulates the result information from a given host
type Result struct {
	Host     string        `json:"host"`
	Matches  []Match       `json:"matches"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error"`
}

// Match type encapsulates the App information from a match on a document
type Match struct {
	App
	AppName string     `json:"app_name"`
	Matches [][]string `json:"matches"`
	Version string     `json:"version"`
}

func (m *Match) updateVersion(version string) {
	if version != "" {
		m.Version = version
	}
}

// Init sets up all the workders, reads in the host data and returns the results channel or an error
func Init(workers int, hosts io.Reader, appsFile string) (chan Result, error) {
	wa, err := NewWebAnalyzer(workers, appsFile)
	if err != nil {
		return nil, err
	}
	// send hosts line by line to worker channel
	go func(hosts io.Reader, wa *WebAnalyzer) {
		scanner := bufio.NewScanner(hosts)
		for scanner.Scan() {
			url := scanner.Text()
			wa.Schedule(NewOnlineJob(url, "", nil))
		}
		// wait for workers to finish, the close result channel to signal finish of scan
		wa.Close()
	}(hosts, wa)
	return wa.Results, nil
}

type WebAnalyzer struct {
	Results chan Result
	jobs    chan *Job
	wg      *sync.WaitGroup
}

// NewWebanalyzer returns an analyzer struct for an ongoing job, which may be
// "fed" jobs via a method and returns them via a channel when complete.
func NewWebAnalyzer(workers int, appsFile string) (*WebAnalyzer, error) {
	wa := new(WebAnalyzer)
	wa.Results = make(chan Result)
	wa.jobs = make(chan *Job)
	wa.wg = new(sync.WaitGroup)
	if err := loadApps(appsFile); err != nil {
		return nil, err
	}
	// start workers
	initWorker(workers, wa.jobs, wa.Results, wa.wg)
	return wa, nil
}

func (wa *WebAnalyzer) Schedule(job *Job) {
	wa.jobs <- job
}

func (wa *WebAnalyzer) Close() {
	close(wa.jobs)
	wa.wg.Wait()
	close(wa.Results)
}
