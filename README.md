# webanalyze

This is a port of [Wappalyzer](https://github.com/AliasIO/Wappalyzer) in Go. This tool is designed to be performant and allows to test huge lists of hosts.

## Installation and usage

    $ go get -u github.com/rverton/webanalyze/...
    $ webanalyze -update # loads new apps.json file from wappalyzer project
    $ webanalyze -h
    Usage of webanalyze:
      -apps string
            app definition file. (default "apps.json")
      -host string
            single host to test
      -hosts string
            filename with hosts, one host per line.
      -output string
            output format (stdout|csv|json) (default "stdout")
      -update
            update apps file
      -worker int
            number of worker (default 4)

The `-update` flags downloads a current version of apps.json from the [wappalyzer repository](https://github.com/AliasIO/Wappalyzer) to the current folder.

## Display

Run `cmd/webanalyze/index.html` (on sth. like SimpleHTTPServer) to display results in a searchable dashboard.

## Development / Usage as a lib

See cmd/webanalyze/main.go for an example.

## Example

    $ webanalyze -host stackshare.io
    2017/06/19 10:22:23 Scanning with 4 workers.
    2017/06/19 10:22:24 [+] http://stackshare.io (556.835509ms):
    2017/06/19 10:22:24 	- Express,  (Web Frameworks, Web Servers)
    2017/06/19 10:22:24 	- Nginx, 1.8.1 (Web Servers)
    2017/06/19 10:22:24 	- Ruby on Rails,  (Web Frameworks)
    2017/06/19 10:22:24 	- Google Font API,  (Font Scripts)

    $ webanalyze -host stackshare.io -output csv
    2017/06/19 10:22:50 Scanning with 4 workers.
    Host,Category,App,Version
    http://stackshare.io,"Web Frameworks,Web Servers",Express,
    http://stackshare.io,Web Servers,Nginx,1.8.1
    http://stackshare.io,Font Scripts,Google Font API,
    http://stackshare.io,Web Frameworks,Ruby on Rails,
