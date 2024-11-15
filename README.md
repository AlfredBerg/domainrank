# Domainrank

Domainrank is a tool to lookup apex domains ranking in the [tranco list](https://tranco-list.eu/). The full list of ~4.7m apex domains is used.  
If a subdomain is passed to the tool the apex domain of the subdomain will be looked up.  
The list is updated the first time the tool is run and then every 7 days (or if the tmp directory is cleaned).  

## Install
```
go install github.com/AlfredBerg/domainrank@latest
```

## Example
```
$ cat domains 
example.com
www.example.com
doesnotexist-1kvjg.com

$ cat domains | domainrank 
example.com example.com 255
www.example.com example.com 255
doesnotexist-1kvjg.com doesnotexist-1kvjg.com -1
```
