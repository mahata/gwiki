FROM alpine:3.4
MAINTAINER "Yasunori Mahata" <mahata777+docker@gmail.com>

ENV GOPATH /root/go-workspace
RUN apk add --no-cache --update go
RUN apk add --no-cache --update alpine-sdk
RUN apk add --no-cache --update git
RUN go get github.com/russross/blackfriday
RUN go get github.com/mahata/gwiki

COPY config.json.tpl /root/go-workspace/src/github.com/mahata/gwiki/config.json
# FixMe: Modify config.json

RUN mkdir -p /usr/local/gwiki/data/img
RUN mkdir -p /usr/local/gwiki/data/txt
RUN touch /usr/local/gwiki/data/txt/index.txt

