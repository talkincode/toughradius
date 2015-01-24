#!/bin/sh

if [ ! -d /var/toughradius ]; then
    mkdir -p /var/toughradius
    echo "mkdir /var/toughradius"
fi

echo "docker run ..."

docker run -d -P -v /var/toughradius:/var/toughradius \
  -p 3306:3306 -p 1812:1812/udp -p 1813:1813/udp \
  -p 1815:1815 -p 1816:1816 \
  --name toughradius talkincode/centos7-toughradius \
  sh /opt/startup.sh