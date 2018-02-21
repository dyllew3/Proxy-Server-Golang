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

## File contents
This sections contains information about each file and what it does

### base.html
File is the management page for the proxy server. This page contains two forms
one which allows the user to add a url to the blocklist and the other which allows the user to remove a url from the blocklist. It also has a link to the "/blocked" which displays the blocklist to the user.

## blocked.json
