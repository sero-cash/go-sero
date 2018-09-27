FROM ubuntu:xenial

ENV PATH=/usr/lib/go-1.9/bin:$PATH

RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends golang-1.9 git make gcc libc-dev ca-certificates && \
  git clone --depth 1 https://github.com/ethereum/go-sero && \
  (cd go-sero && make gero) && \
  cp go-sero/build/bin/gero /gero && \
  apt-get remove -y golang-1.9 git make gcc libc-dev && apt autoremove -y && apt-get clean && \
  rm -rf /go-sero

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gero"]
