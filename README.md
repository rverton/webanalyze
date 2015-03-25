### Wappalyze

This is a port of [Wappalyzer](https://github.com/AliasIO/Wappalyzer) in Go. This tool is designed to be performant and allows to test huge lists of hosts. 

## Usage

    $ ./webanalyze -h
    Usage of ./webanalyze:
      -apps="apps.json": app definition file.
      -host="": Single host to test
      -hosts="filename": List of hosts. One line per host.
      -update=false: Update apps file
      -worker=50: Number of worker.

The -update flags downloads a current version of apps.json from the wappalyzer repository.
