FROM golang:1.24-bookworm AS build

WORKDIR /src/app

RUN apt-get update \
  && apt-get -y install netcat-openbsd ca-certificates \
  && apt-get clean

COPY go.* ./
RUN go mod download
COPY . . 

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 go build -o monitor .

FROM scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=build /src/app/db/migrations/ /app/db/migrations/
COPY --from=build /src/app/web/static/ /app/web/static/
COPY --from=build /src/app/web/views/ /app/web/views/
COPY --from=build /src/app/monitor /app/monitor
COPY --from=build /src/app/relays.txt /app/relays.txt

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

CMD ["/app/monitor"]
