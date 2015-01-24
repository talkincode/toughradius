#!/bin/sh

if [ ! -d /var/toughradius ]; then
    mkdir -p /var/toughradius
    echo "mkdir /var/toughradius"
fi

echo "docker run ..."

docker run -d -v /var/toughradius:/var/toughradius \
  --name toughradius talkincode/centos7-toughradius \
  sh /opt/startup.sh