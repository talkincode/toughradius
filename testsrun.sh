#!/bin/bash

python toughctl --initdb -c toughradius/tests/test.json

python coverage run toughctl --standalone -exitwith 60 -c toughradius/tests/test.json &

sleep 10

python trial toughradius.tests

sleep 50
