#!/bin/sh

if [ ! -f "/var/toughradius/data" ];then
    mkdir -p /var/toughradius/data
fi

if [ ! -f "/var/toughradius/.install" ];then
    pypy /opt/toughradius/toughctl --initdb
    echo "ok" > /var/toughradius/.install
    echo "init database ok!"
fi

exec "$@"