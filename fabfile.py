#!/usr/bin/env python
import sys,os,time
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
from toughradius import __version__


def tag():
    local("git tag -a v%s -m 'version %s'"%(__version__,__version__))
    local("git push origin v%s:v%s"%(__version__,__version__))


def tag2():
    local("git tag -a v%s -m 'version %s'"%(__version__,__version__))
    local("git push src v%s:v%s"%(__version__,__version__))

def tests(exitwith=3600.0,kill=1):
    if kill:
        try:
            local("ps aux | grep '/test.json' | awk '{print $2}' | xargs  kill")
            local("rm -f /tmp/trtest.log ")
        except:
            pass
    print '\n-------------- INIT TEST DB ---------------------------- \n'
    local("pypy toughctl --initdb -c toughradius/tests/test.json")
    print '\n-------------- RUN TESTING SERVER ---------------------------- \n'
    local("pypy coverage run toughctl --standalone -exitwith %s -c toughradius/tests/test.json > /tmp/trtest.log &" % exitwith)
    while 1:
        time.sleep(1)
        print "waiting server start..."
        if open("/tmp/trtest.log").read().find("testing application running") >= 0:
            break
    print '\n-------------- RUN TESTS ---------------------------- \n'
    local("pypy coverage run trial toughradius.tests")

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
    local("git push src master")



def auth():
    local("pypy toughctl --auth -c ~/toughradius_test.json")

def acct():
    local("pypy toughctl --acct -c ~/toughradius_test.json")

def worker():
    local("pypy toughctl --worker -c ~/toughradius_test.json")


def manage():
    local("pypy toughctl --manage -c ~/toughradius_test.json")

def task():
    local("pypy toughctl --task -c ~/toughradius_test.json")

def all():
    local("pypy toughctl --standalone -c ~/toughradius_test.json")

def run():
    local("pypy toughctl --run -c ~/toughradius_test.json")

def super():
    local("supervisord -c etc/supervisord_test.conf")

def initdb():
    local("pypy toughctl --initdb -c ~/toughradius_test.json")
    local("pypy toughctl --inittest -c ~/toughradius_test.json")


def uplib():
    local("pypy -m pip install https://github.com/talkincode/toughlib/archive/master.zip --upgrade --no-deps")

def uplib2():
    local("pypy -m pip install https://github.com/talkincode/txradius/archive/master.zip --upgrade --no-deps")
