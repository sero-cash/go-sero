FROM ubuntu:xenial
ENV LD_LIBRARY_PATH=/geropkg/czero/lib
COPY rungero.docker /usr/local/bin/rungero
COPY geroConfig.toml /usr/local/bin/geroConfig.toml
RUN \
  apt-get update && apt-get upgrade -q -y && \
  apt-get install -y --no-install-recommends wget make gcc libc-dev libgmp-dev ca-certificates && \
  wget https://sero-media-1256272584.cos.ap-shanghai.myqcloud.com/gero/v1.0.0-rc6/gero-v1.0.0-rc6-linux-amd64-v4.tar.gz &&\
  tar -zxvf gero-v1.0.0-rc6-linux-amd64-v4.tar.gz  && \
  cp geropkg/bin/gero /gero

EXPOSE 8545
EXPOSE 53717

ENTRYPOINT ["/usr/local/bin/rungero"]
