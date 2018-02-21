# Proxy-Server-Golang
This repo contains a basic proxy server written in Golang allows for

* HTTP/HTTPS.
* Dynamically blocking URLs.
* Management page.
* Caching http pages.

Repo contains 4 files

1. base.html
2. blocked.json
3. cache.go
4. server.go

In order to run the program on windows enter this command in this directory
```
go run server.go cache.go
```

In linux you can run the program using command above or use this command

```
go run *.go
```
If you wish to create an executable that runs the server use following command
```
go build server.go cache.go
```

## File contents
This sections contains information about each file and what it does

### base.html
File is the management page for the proxy server. This page contains two forms
one which allows the user to add a url to the blocklist and the other which allows the user to remove a url from the blocklist. It also has a link to the "/blocked" which displays the blocklist to the user.

### blocked.json
This contains the current version of the blocklist in json format like so
```json
[{"URL":"example.com"}, {"URL":"anotherone.co.uk"}]
```
The json maps the string key **"URL"** to **a string value**
This file is written to by server.go

### cache.go
Implements the cache system for the server. The cache maps a URL string to a byte array, this byte array represents a http response received after a request has been sent. When the http handler function gets a request it checks if request url is in cache, if it is then the byte array is returned, if not then the request is sent to the destination and the response is converted to a byte array and stored in the cache. However if the appropriate byte array is in the cache it checks if it has become stale(no longer a valid response), if it's stale then it passes the request to the destination and stores the new response.
