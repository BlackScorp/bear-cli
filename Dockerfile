FROM alpine:latest

RUN apk add --no-cache \
  go \
  git \
  ripgrep \
  fd \
  jq \
  curl \
  tree \
  bat \
  htmlq \
  build-base


WORKDIR /app

ENV GO111MODULE=on
ENV CGO_ENABLED=0

COPY . .

RUN go mod init baer && go mod tidy && go build -o /usr/local/bin/baer-cli ./src

WORKDIR /project

ENTRYPOINT ["baer-cli"]