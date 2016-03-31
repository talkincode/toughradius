#!/usr/bin/env python
import sys,os,time
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
from toughradius import __version__


def tag():
    local("git tag -a v%s -m 'version %s'"%(__version__,__version__))
    local("git push origin v%s:v%s"%(__version__,__version__))

def commit():
    try:
        local("ps aux | grep '/test.json' | awk '{print $2}' | xargs  kill")
    except:
        pass
    local("echo 'coverage report: version:%s   date:%s' > coverage.txt" % (__version__,time.ctime()))
    local("echo >> coverage.txt")
    local("coverage report >> coverage.txt")
    local("git status && git add .")
    local("git commit -m \"%s\"" % raw_input("type message:"))
    local("git push origin master")


def all():
    local("venv/bin/python radiusctl standalone -c ~/toughradius_test.json")


def initdb():
    local("venv/bin/python radiusctl initdb -c ~/toughradius_test.json")
    local("venv/bin/python radiusctl inittest -c ~/toughradius_test.json")

