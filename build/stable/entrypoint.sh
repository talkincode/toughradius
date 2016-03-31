#!/bin/sh

if [ ! -f "/var/toughradius/data" ];then
    mkdir -p /var/toughradius/data
fi

if [ ! -f "/var/toughradius/.install" ];then
    pypy /opt/toughradius/radiusctl initdb -c /etc/toughradius.json
    echo "ok" > /var/toughradius/.install
    echo "init database ok!"
fi

exec "$@"