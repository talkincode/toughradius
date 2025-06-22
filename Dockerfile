FROM golang:1.20.0-buster AS builder

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags \
     '-s -w -extldflags "-static"'  -o /toughradius main.go

FROM alpine:3.19

RUN apk add --no-cache curl postgresql14-client

COPY --from=builder /toughradius /usr/local/bin/toughradius

RUN chmod +x /usr/local/bin/toughradius

EXPOSE 1816 1817 1818 1819 1812/tcp 1812/udp 1813/udp

ENTRYPOINT ["/usr/local/bin/toughradius"]