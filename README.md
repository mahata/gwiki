# Gwiki

[![Circle CI](https://circleci.com/gh/mahata/gwiki.svg?style=svg)](https://circleci.com/gh/mahata/gwiki)

Markdown Wiki implementation in Go.

## Disclaimer

You shouldn't use this yet. It's still a PoC implementation.

## Install using docker-compose

```
$ sudo mkdir -p /usr/local/gwiki/data
$ sudo touch /usr/local/gwiki/data/index.txt
$ docker-compose up -d
```

## Build Docker Image

```
$ docker build -t mahata/gwiki .
$ docker push mahata/gwiki
```

