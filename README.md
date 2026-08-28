# webanalyze

This is a port of Wappalyzer in Go. This tool is designed to be performant and allows to test huge lists of hosts.

> [!NOTE]
> Because Wappalyzer removed the public access to their app definitions, webanalyze currently loads definitions from [enthec](https://github.com/enthec/webappanalyzer).

## Installation and usage


### Precompiled releases
Precompiled releases can be downloaded directly [here](https://github.com/rverton/webanalyze/releases).

### Build
If you want to build for yourself:

    $ go install -v github.com/rverton/webanalyze/cmd/webanalyze@latest

## Usage

    $ webanalyze -update # loads new technologies.json file from wappalyzer project
    $ webanalyze -h
    Usage of webanalyze:
      -apps string
            app definition file. (default "technologies.json")
      -crawl int
            links to follow from the root page (default 0)
      -header value
            custom HTTP request header, repeatable (e.g. 'User-Agent: webanalyze')
      -host string
            single host to test
      -hosts string
            filename with hosts, one host per line.
      -output string
            output format (stdout|csv|json) (default "stdout")
      -search
            searches all urls with same base domain (i.e. example.com and sub.example.com) (default true)
      -silent
    	    avoid printing header (default false)
      -update
            update apps file
      -worker int
            number of worker (default 4)

The `-update` flags downloads a current version of `technologies.json` from [enthec](https://github.com/enthec/webappanalyzer).

Custom request headers can be supplied more than once:

    $ webanalyze -host example.com -header "User-Agent: Mozilla/5.0" -header "Authorization: Bearer token"

### Docker

```bash
# Clone the repo
git clone https://github.com/rverton/webanalyze.git
# Build the container
docker build -t webanalyze:latest webanalyze
# Run the container
docker run -it webanalyze:latest -h
```

## Development / Usage as a lib

See `cmd/webanalyze/main.go` for an example on how to use this as a library.

## Example

    $ ./webanalyze -host robinverton.de -crawl 1
     :: webanalyze        : v1.0
     :: workers           : 4
     :: apps              : technologies.json
     :: crawl count       : 1
     :: search subdomains : true

    https://robinverton.de/hire/ (0.5s):
        Highlight.js,  (Miscellaneous)
        Netlify,  (Web Servers, CDN)
        Google Font API,  (Font Scripts)
    http://robinverton.de (0.8s):
        Highlight.js,  (Miscellaneous)
        Netlify,  (Web Servers, CDN)
        Hugo, 0.42.1 (Static Site Generator)
        Google Font API,  (Font Scripts)

    $ ./webanalyze -host robinverton.de -crawl 1 -output csv
     :: webanalyze        : v1.0
     :: workers           : 4
     :: apps              : technologies.json
     :: crawl count       : 1
     :: search subdomains : true

    Host,Category,App,Version
    https://robinverton.de/hire/,Miscellaneous,Highlight.js,
    https://robinverton.de/hire/,Font Scripts,Google Font API,
    https://robinverton.de/hire/,"Web Servers,CDN",Netlify,
    http://robinverton.de,"Web Servers,CDN",Netlify,
    http://robinverton.de,Static Site Generator,Hugo,0.42.1
    http://robinverton.de,Miscellaneous,Highlight.js,
    http://robinverton.de,Font Scripts,Google Font API,
