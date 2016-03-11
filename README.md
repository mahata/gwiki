# Gwiki

[![Circle CI](https://circleci.com/gh/mahata/gwiki.svg?style=svg)](https://circleci.com/gh/mahata/gwiki)

Markdown Wiki implementation in Go.

## Disclaimer

You shouldn't use this yet. It's still a PoC implementation.

## Install using Docker

It works, but it's apparently not the best way to run it. Still in progress.

```
$ docker pull mahata/gwiki
$ docker run -p 8080:8080 -ti mahata/gwiki /bin/sh -c "cd /usr/local/gwiki; go run wiki.go"

(You need to create "/usr/local/gwiki-data/index.txt" beforehand - it's ok for index.txt to be empty)
$ docker run -p 8080:8080 -v /usr/local/gwiki-data:/usr/local/gwiki/data -ti mahata/gwiki /bin/sh -c "cd /usr/local/gwiki; go run wiki.go"
```
