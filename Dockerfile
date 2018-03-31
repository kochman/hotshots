FROM golang:1.10

RUN go get github.com/kardianos/govendor
RUN mkdir -p /go/src/github.com/kochman/hotshots
WORKDIR /go/src/github.com/kochman/hotshots
COPY vendor/vendor.json ./vendor/
RUN govendor sync

COPY . /go/src/github.com/kochman/hotshots
RUN go install github.com/kochman/hotshots

ENV HOTSHOTS_LISTEN_URL 0.0.0.0:8000
VOLUME /var/hotshots
EXPOSE 8000
CMD ["/go/bin/hotshots", "server"]
