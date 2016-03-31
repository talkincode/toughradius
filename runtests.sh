#!/bin/sh

venv/bin/python radiusctl initdb -c toughradius/tests/test.json

venv/bin/coverage run radiusctl standalone --exitwith 60 -c toughradius/tests/test.json > /tmp/trtest.log &

echo "wait 10 second.."

sleep 10

echo "starting test.."

venv/bin/trial toughradius.tests

cat /tmp/trtest.log

