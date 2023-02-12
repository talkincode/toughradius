FROM python:3.9.16-alpine3.17
RUN apk add --no-cache curl
RUN apk add --no-cache postgresql14-client
WORKDIR /var/toughradius

COPY ./release/toughradius /usr/bin/toughradius
RUN chmod +x /usr/bin/toughradius

EXPOSE 1816 1817 1818 1819 1812/tcp 1812/udp 1813/udp

CMD ["/usr/bin/toughradius"]