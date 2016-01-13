#!/usr/bin/env python
import sys,os
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
from toughradius import __version__


def tag():
    local("git tag -a v%s -m 'version %s'"%(__version__,__version__))
    local("git push origin v%s:v%s"%(__version__,__version__))

def auth():
    local("pypy toughctl --auth -c toughradius.json")


def acct():
    local("pypy toughctl --acct -c toughradius.json")


def manage():
    local("pypy toughctl --manage -c toughradius.json")


def all():
    local("pypy toughctl --standalone -c toughradius.json")


def initdb():
    local("pypy toughctl --initdb -c toughradius.json")

def uplib():
    local("pypy -m pip install https://github.com/talkincode/toughlib/archive/master.zip --upgrade --no-deps")

def uplib2():
    local("pypy -m pip install https://github.com/talkincode/txradius/archive/master.zip --upgrade --no-deps")
