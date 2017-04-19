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

    $ webanalyze -host stackshare.io
    2017/04/19 16:21:18 Scanning with 4 workers.
    2017/04/19 16:21:19 [+] http://stackshare.io (616.713344ms):
    2017/04/19 16:21:19 	- Ruby on Rails (Web Frameworks)
    2017/04/19 16:21:19 	- Google Font API (Font Scripts)
    2017/04/19 16:21:19 	- Express (Web Frameworks, Web Servers)
    2017/04/19 16:21:19 	- Nginx (Web Servers)
