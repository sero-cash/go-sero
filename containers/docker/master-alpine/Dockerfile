FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/ethereum/go-sero && \
  (cd go-sero && make gero) && \
  cp go-sero/build/bin/gero /gero && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-sero && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gero"]
