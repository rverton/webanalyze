# webanalyze

This is a port of [Wappalyzer](https://github.com/AliasIO/Wappalyzer) in Go. This tool is designed to be performant and allows to test huge lists of hosts.

## Installation and usage

    $ go get -u github.com/rverton/webanalyze/...
    $ webanalyze -update # loads new apps.json file from wappalyzer project
    $ webanalyze -h
    Usage of webanalyze:
      -apps string
            app definition file. (default "apps.json")
      -crawl int
            links to follow from the root page (default 0)
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

See `cmd/webanalyze/main.go` for an example.

## Example

    $ webanalyze -host https://stackshare.io
    2019/01/05 23:41:45 Scanning with 4 workers.
    2019/01/05 23:41:46 [+] https://stackshare.io (1.025640074s):
    2019/01/05 23:41:46 	- jQuery,  (JavaScript Libraries)
    2019/01/05 23:41:46 	- Cowboy,  (Web Frameworks, Web Servers)
    2019/01/05 23:41:46 	- Erlang,  (Programming Languages)
    2019/01/05 23:41:46 	- Ruby on Rails,  (Web Frameworks)
    2019/01/05 23:41:46 	- Ruby,  (Programming Languages)
    
    $ webanalyze -host https://stackshare.io -output csv
    2019/01/05 23:45:04 Scanning with 4 workers.
    Host,Category,App,Version
    https://stackshare.io,"Web Frameworks,Web Servers",Cowboy,
    https://stackshare.io,Programming Languages,Erlang,
    https://stackshare.io,Web Frameworks,Ruby on Rails,
    https://stackshare.io,Programming Languages,Ruby,
    https://stackshare.io,JavaScript Libraries,jQuery,
