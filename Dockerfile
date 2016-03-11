FROM mahata/golpine:3.3
MAINTAINER "Yasunori Mahata" <mahata777+docker@gmail.com>


# COPY gwiki_linux_amd64 /usr/local/gwiki/wiki  # This doesn't work somehow (TдT); Following 4 lines are just a dirty workaround
ENV GOPATH /root/go-workspace
COPY wiki.go /usr/local/gwiki/wiki.go
RUN apk add --no-cache --update git
RUN go get github.com/russross/blackfriday

COPY config.json /usr/local/gwiki/config.json
COPY edit.html /usr/local/gwiki/edit.html
COPY view.html /usr/local/gwiki/view.html

RUN mkdir /usr/local/gwiki/data
RUN touch /usr/local/gwiki/data/index.txt
