#!/bin/bash

python toughctl --initdb -c toughradius/tests/test.json

python coverage run toughctl --standalone -exitwith 120 -c toughradius/tests/test.json &

sleep 15

python trial toughradius.tests


