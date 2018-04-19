
FROM node:9 as js-build
WORKDIR /root
COPY web /root
RUN npm install
RUN npm run build

FROM golang:1.10 as go-build
RUN apt-get update \
    && apt-get install -y libgphoto2-dev \
    && rm -rf /var/lib/apt/lists/*
RUN go get github.com/kardianos/govendor
RUN mkdir -p /go/src/github.com/kochman/hotshots
WORKDIR /go/src/github.com/kochman/hotshots
COPY vendor/vendor.json ./vendor/
RUN govendor sync
COPY . /go/src/github.com/kochman/hotshots
RUN go build

FROM debian:stable as runner
RUN apt-get update \
    && apt-get install -y libgphoto2-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /root
COPY --from=go-build /go/src/github.com/kochman/hotshots/hotshots .
COPY --from=js-build /root/assets ./web/assets
COPY --from=js-build /root/*.html ./web/
ENV HOTSHOTS_LISTEN_URL 0.0.0.0:8000
VOLUME /var/hotshots
EXPOSE 8000/tcp
CMD ["/root/hotshots", "server"]
