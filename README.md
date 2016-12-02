# webanalyze

This is a port of [Wappalyzer](https://github.com/AliasIO/Wappalyzer) in Go. This tool is designed to be performant and allows to test huge lists of hosts.

## Installation and usage

    $ go get -u github.com/rverton/webanalyze/...
    $ webanalyze -update
    $ webanalyze -h
    Usage of ./webanalyze:
      -apps string
            app definition file. (default "apps.json")
      -csv string
            export to csv file
      -host string
            single host to test
      -hosts string
            list of hosts to test, one host per line.
      -json string
            output to json file
      -update
            update apps file
      -worker int
            number of worker (default 4)

The `-update` flags downloads a current version of apps.json from the [wappalyzer repository](https://github.com/AliasIO/Wappalyzer) to the current folder.

## Display

Run cmd/webanalyze/index.html (on sth. like SimpleHTTPServer) to display results in a searchable dashboard.

## Development / Usage as a lib

See cmd/webanalyze/main.go for an example.

## Example

    $ webanalyze -host="http://stackshare.io"
    2015/05/12 09:27:12 Loaded 752 app definitions
    2015/05/12 09:27:12 Scanning with 4 workers.
    [+] http://stackshare.io (1.722697135s):
        - Glyphicons
        - Google Font API
        - RequireJS
        - Font Awesome
        - Prototype
        - jQuery
