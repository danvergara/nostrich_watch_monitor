FROM golang:1.24-bookworm AS builder

WORKDIR /src/app

# install system dependencies
RUN apt-get update \
  && apt-get -y install netcat-openbsd \
  && apt-get clean

COPY go.* ./
RUN go mod download
COPY . . 

ARG TARGETOS
ARG TARGETARCH

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o monitor .

FROM scratch AS bin

LABEL org.opencontainers.image.documentation="https://github.com/danvergara/nostrich_watch_monitor" \
	org.opencontainers.image.source="https://github.com/danvergara/nostrich_watch_monitor" \
	org.opencontainers.image.title="nostrich_watch_monitor"

COPY --from=builder /src/app/monitor /bin/monitor
