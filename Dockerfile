
FROM alpine:latest AS builder

RUN apk add --no-cache \
  go \
  build-base

WORKDIR /app

ENV GO111MODULE=on
ENV CGO_ENABLED=0

COPY . .

RUN go mod init baer || true
RUN go mod tidy
RUN go build -o baer-cli ./src


FROM alpine:latest

RUN apk add --no-cache \
  git \
  ripgrep \
  fd \
  jq \
  curl \
  tree \
  bat \
  htmlq

WORKDIR /project

COPY --from=builder /app/baer-cli /usr/local/bin/baer-cli

ENTRYPOINT ["baer-cli"]