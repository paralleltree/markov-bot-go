FROM alpine AS build-mecab
WORKDIR /build
RUN mkdir /app
RUN apk update && apk --no-cache add git make gcc g++ automake autoconf \
  bash gnu-libiconv xz curl patch file openssl perl
RUN git clone --depth=1 https://github.com/taku910/mecab && (cd mecab/mecab && ./configure && make && make install) && rm -rf mecab
RUN git clone --depth=1 https://github.com/neologd/mecab-ipadic-neologd && (cd mecab-ipadic-neologd && sed -i 's/2,2/2/' libexec/make-mecab-ipadic-neologd.sh && ./bin/install-mecab-ipadic-neologd -n -y -u)

FROM golang:1.18.2-alpine AS build-cli
ADD . /src
WORKDIR /src
RUN GOOS=linux go build -o bot cmd/cli/main.go

FROM alpine
WORKDIR /app
RUN apk --no-cache add libstdc++ libgcc
COPY --from=build-mecab /usr/local /usr/local
COPY --from=build-cli /src/bot /app
