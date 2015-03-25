package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const WAPPALYZER_URL = "https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/apps.json"

type StringArray []string

type App struct {
	Cats    []int             `json:"cats"`
	Headers map[string]string `json:"headers"`
	HTML    StringArray       `json:"html"`
	Script  StringArray       `json:"script"`
	URL     StringArray       `json:"url"`
	Website string            `json:"website"`

	HTMLRegex   []*regexp.Regexp
	ScriptRegex []*regexp.Regexp
	URLRegex    []*regexp.Regexp
	HeaderRegex map[string]*regexp.Regexp `json:"headers"`
}

type AppsDefinition struct {
	Apps map[string]App `json:"apps"`
}

type Match struct {
	AppName    string     `json:"app"`
	AppWebsite string     `json:"app_website"`
	Matches    [][]string `json:"matches"`
}

// custom unmarshaler for handling bogus apps.json types from wappalyzer
func (t *StringArray) UnmarshalJSON(data []byte) error {
	var s string
	var sa []string

	if err := json.Unmarshal(data, &s); err != nil {

		// not a string, so maybe []string?
		if err := json.Unmarshal(data, &sa); err != nil {
			return err
		}
		*t = sa
		return nil
	}
	*t = StringArray{s}
	return nil
}

func updateApps(url string) error {
	return downloadFile(url, WAPPALYZER_URL)
}

func downloadFile(from, to string) error {

	resp, err := http.Get(from)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	return err
}

// load apps from file
func LoadApps(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)
	if err = dec.Decode(&appDefs); err != nil {
		return err
	}

	// compile regular expressions
	for key, value := range appDefs.Apps {

		app := appDefs.Apps[key]
		app.HTMLRegex = compileRegexes(value.HTML)
		app.ScriptRegex = compileRegexes(value.Script)
		app.URLRegex = compileRegexes(value.URL)

		for key, value := range app.Headers {
			app.HeaderRegex[key], err = regexp.Compile(value)
			if err != nil {
				// ignore failed compiling for now
				// log.Printf("waring: compiling regex for header failed: %v", err)
			}
		}

		appDefs.Apps[key] = app

	}

	return nil
}

func compileRegexes(s StringArray) []*regexp.Regexp {
	var list []*regexp.Regexp

	for _, regexString := range s {

		// Filter out webapplyzer attributes from regular expression
		cleaned := strings.Split(regexString, "\\;")[0]

		regex, err := regexp.Compile(cleaned)
		if err != nil {
			// ignore failed compiling for now
			// log.Printf("warning: compiling regexp for failed: %v", regexString, err)
		} else {
			list = append(list, regex)
		}
	}

	return list
}
