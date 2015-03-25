# webanalyze

This is a port of [Wappalyzer](https://github.com/AliasIO/Wappalyzer) in Go. This tool is designed to be performant and allows to test huge lists of hosts. 

## Installation and usage

    $ go install github.com/rverton/webanalyze
    $ webanalyze -update
    $ webanalyze -h
    Usage of ./webanalyze:
      -apps="apps.json": app definition file.
      -host="": Single host to test
      -hosts="filename": List of hosts. One line per url.
      -update=false: Update apps file
      -worker=50: Number of worker.

The -update flags downloads a current version of apps.json from the wappalyzer repository.

## Use as a library

See the main\_test.go file for an example on how to integrate.
