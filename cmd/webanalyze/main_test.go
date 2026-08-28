package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/rverton/webanalyze"
)

func TestAddRequestHeader(t *testing.T) {
	headers := make(http.Header)

	if err := addRequestHeader(headers, "Authorization: Bearer abc:def"); err != nil {
		t.Fatalf("addRequestHeader returned an error: %v", err)
	}
	if err := addRequestHeader(headers, "X-Custom: first"); err != nil {
		t.Fatalf("addRequestHeader returned an error: %v", err)
	}
	if err := addRequestHeader(headers, "X-Custom: second"); err != nil {
		t.Fatalf("addRequestHeader returned an error: %v", err)
	}

	if got, want := headers.Get("Authorization"), "Bearer abc:def"; got != want {
		t.Errorf("got Authorization header %q, want %q", got, want)
	}
	if got, want := headers.Values("X-Custom"), []string{"first", "second"}; !slices.Equal(got, want) {
		t.Errorf("got X-Custom headers %q, want %q", got, want)
	}
}

func TestAddRequestHeaderRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"invalid", ": value"} {
		t.Run(value, func(t *testing.T) {
			if err := addRequestHeader(make(http.Header), value); err == nil {
				t.Fatalf("addRequestHeader(%q) returned no error", value)
			}
		})
	}
}

func TestConcurrentCSVOutput(t *testing.T) {
	const (
		resultCount = 100
		matchCount  = 5
	)

	previousOutputMethod := outputMethod
	outputMethod = "csv"
	t.Cleanup(func() {
		outputMethod = previousOutputMethod
	})

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	var wg sync.WaitGroup
	for i := 0; i < resultCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			matches := make([]webanalyze.Match, matchCount)
			for j := range matches {
				matches[j] = webanalyze.Match{
					App: webanalyze.App{
						CatNames: []string{"Category A", "Category B"},
					},
					AppName: fmt.Sprintf("app-%d", j),
					Version: "1.0",
				}
			}

			output(webanalyze.Result{
				Host:    fmt.Sprintf("http://example-%d.com", i),
				Matches: matches,
			}, nil, writer)
		}(i)
	}
	wg.Wait()

	records, err := csv.NewReader(bytes.NewReader(buffer.Bytes())).ReadAll()
	if err != nil {
		t.Fatalf("cannot parse CSV output: %v", err)
	}

	if got, want := len(records), resultCount*matchCount; got != want {
		t.Fatalf("got %d CSV records, want %d", got, want)
	}

	for i, record := range records {
		if got, want := len(record), 4; got != want {
			t.Fatalf("record %d has %d fields, want %d: %q", i, got, want, record)
		}
	}
}
