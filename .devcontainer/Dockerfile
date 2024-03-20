FROM debian:bookworm-slim AS build-mecab
WORKDIR /build
RUN mkdir /app
RUN apt-get update && apt-get -y install git make gcc g++ build-essential libc6-dev automake autoconf \
  xz-utils curl patch file openssl perl
RUN git clone --depth=1 https://github.com/taku910/mecab && (cd mecab/mecab && ./configure && make && make install) && rm -rf mecab && ldconfig
RUN git clone --depth=1 https://github.com/neologd/mecab-ipadic-neologd && (cd mecab-ipadic-neologd && sed -i -e 's/2,2/2/' -e 's/ORG_DIC_URL_LIST\[2\]/ORG_DIC_URL_LIST[0]/' -e 's/DIST_SITE_URL_LIST\[2\]/DIST_SITE_URL_LIST[0]/' libexec/make-mecab-ipadic-neologd.sh && ./bin/install-mecab-ipadic-neologd -n -y -u)

FROM golang:1.22-bookworm
WORKDIR /src
COPY --from=build-mecab /usr/local /usr/local
RUN ldconfig
COPY . .
