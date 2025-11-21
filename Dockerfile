FROM golang:1.24-bookworm AS builder

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags \
     '-s -w -extldflags "-static"'  -o /toughradius main.go

FROM alpine:3.19

RUN apk add --no-cache curl

COPY --from=builder /toughradius /usr/local/bin/toughradius

RUN chmod +x /usr/local/bin/toughradius

# Expose required ports:
# 1816 - Web/Admin API (HTTP)
# 1812 - RADIUS Authentication (UDP)
# 1813 - RADIUS Accounting (UDP)
# 2083 - RadSec (RADIUS over TLS)
EXPOSE 1816/tcp 1812/udp 1813/udp 2083/tcp

ENTRYPOINT ["/usr/local/bin/toughradius"]