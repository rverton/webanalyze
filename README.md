# webanalyze

This is a port of [Wappalyzer](https://github.com/AliasIO/Wappalyzer) in Go. This tool is designed to be performant and allows to test huge lists of hosts. 

## Installation and usage

    $ go get -u github.com/rverton/webanalyze/...
    $ webanalyze -update
    $ webanalyze -h
    Usage of ./webanalyze:
      -apps="apps.json": app definition file.
      -host="": Single host to test
      -hosts="filename": List of hosts. One line per url.
      -update=false: Update apps file
      -worker=4: Number of worker.

The -update flags downloads a current version of apps.json from the wappalyzer repository in the current folder.

## Development

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
