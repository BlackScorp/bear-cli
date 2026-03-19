FROM alpine:latest

# Basis-Tools + Go + deine CLI dependencies
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

# Arbeitsverzeichnis
WORKDIR /app

# Optional: Go ENV setzen
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Source Code rein (für initialen Stand)
COPY . .

# Binary initial bauen (kannst du später überschreiben)
RUN go build -o baer-cli ./src/main.go

RUN ls -la /app
WORKDIR /project

# Default: interaktive Shell ODER direkt CLI starten
ENTRYPOINT ["/app/baer-cli"]