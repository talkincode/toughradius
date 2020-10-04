#!/bin/sh
set -e

test -d /var/toughradius || mkdir -p /var/toughradius

# else default to run whatever the user wanted like "bash" or "sh"
exec "$@"