FROM mahata/golpine:3.3
MAINTAINER "Yasunori Mahata" <mahata777+docker@gmail.com>

ENV GOPATH /root/go-workspace
RUN apk add --no-cache --update git
RUN go get github.com/russross/blackfriday
RUN go get github.com/mahata/gwiki

COPY config.json /root/go-workspace/src/github.com/mahata/gwiki/config.json

RUN mkdir -p /usr/local/gwiki/data/img
RUN mkdir -p /usr/local/gwiki/data/txt
RUN touch /usr/local/gwiki/data/txt/index.txt

