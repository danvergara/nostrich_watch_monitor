FROM golang:1.24-bookworm AS build

WORKDIR /src/app

RUN apt-get update \
  && apt-get -y install netcat-openbsd \
  && apt-get clean

COPY go.* ./
RUN go mod download
COPY . . 

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 go build -o monitor .

FROM scratch

COPY --from=build /src/app/monitor /bin/monitor

CMD ["/bin/monitor"]
