# webanalyze

This is a fork  of [Wappalyzer](https://github.com/rverton/webanalyze/releases) in Go. This tool is designed to be performant and allows to test huge lists of hosts.

👾 Added piping capabilities where the output is a one line csv for each host

## Installation and usage


### Precompiled releases
Precompiled releases can be downloaded directly [here](https://github.com/rverton/webanalyze/releases).

### Build
If you want to build for yourself:

    $ go install -v github.com/TheeEclipse/webanalyze/cmd/webanalyze@latest
    $ webanalyze -update # loads new technologies.json file from wappalyzer project
    Or rather just wget the technologies file
    $ webanalyze -h
    Usage of webanalyze:
      -apps string
            app definition file. (default "technologies.json")
      -crawl int
            links to follow from the root page (default 0)
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


The `-update` flags downloads a current version of `technologies.json` from the [wappalyzer repository](https://github.com/AliasIO/Wappalyzer) to the current folder.


See `cmd/webanalyze/main.go` for an example on how to use this as a library.

## Example

    $ root@Normal-Use-Instance:~# webanalyze -host robinverton.de -crawl 1 -silent
```http://robinverton.de (0.5s): React,  (JavaScript frameworks) HSTS,  (Security) Netlify,  (PaaS, CDN)```

    $ root@Normal-Use-Instance:~# webanalyze -host robinverton.de -crawl 1 -silent | anew 2.txt
```
http://robinverton.de (0.5s): HSTS,  (Security) Netlify,  (PaaS, CDN) React,  (JavaScript frameworks)
root@Normal-Use-Instance:~#
```

    $ root@Normal-Use-Instance:~# webanalyze -host robinverton.de -crawl 1
```
 :: webanalyze        : v0.3.9
 :: workers           : 4
 :: technologies      : technologies.json
 :: crawl count       : 1
 :: search subdomains : true
 :: follow redirects  : false

http://robinverton.de (0.5s): React,  (JavaScript frameworks) HSTS,  (Security) Netlify,  (PaaS, CDN)
root@Normal-Use-Instance:~#
```

