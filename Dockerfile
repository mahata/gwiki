FROM mahata/golpine:3.3
MAINTAINER "Yasunori Mahata" <mahata777+docker@gmail.com>

ENV GOPATH /root/go-workspace
COPY main.go /root/go-workspace/src/github.com/mahata/gwiki/main.go
COPY wiki/wiki.go /root/go-workspace/src/github.com/mahata/gwiki/wiki/wiki.go
RUN apk add --no-cache --update git
RUN go get github.com/russross/blackfriday

COPY config.json /root/go-workspace/src/github.com/mahata/gwiki/config.json
COPY edit.html /root/go-workspace/src/github.com/mahata/gwiki/edit.html
COPY view.html /root/go-workspace/src/github.com/mahata/gwiki/view.html
COPY login.html /root/go-workspace/src/github.com/mahata/gwiki/login.html

RUN mkdir -p /usr/local/gwiki/data
RUN touch /usr/local/gwiki/data/index.txt
