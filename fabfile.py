#!/usr/bin/env python
import sys,os,time
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
from toughradius import __version__
import datetime

env.user = 'root'
env.hosts = ['www.toughstruct.net']
currtime = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

def push():
    message = raw_input("commit msg:")
    local("git add .")
    try:
        local("git commit -m \'%s - %s: %s\'"%(__version__, currtime, message))
    except:
        print 'no commit'
    local("git push origin master")
    local("git push coding master")
    local("git push osc master")

def tag():
    local("git tag -a v%s -m 'version %s'"%(__version__,__version__))
    local("git push origin v%s:v%s"%(__version__,__version__))



def push_stable():
    message = raw_input("commit msg:")
    local("git add .")
    try:
        local("git commit -m \'%s\'"%message)
    except:
        print 'no commit'
    local("git push origin master")
    local("git checkout release-dev")
    local("git merge master --no-ff")
    local("git push origin release-dev")
    local("git checkout release-stable")
    local("git merge release-dev --no-ff")
    local("git push origin release-stable")
    local("git checkout master")

def push_stable2():
    message = raw_input("commit msg:")
    local("git add .")
    try:
        local("git commit -m \'%s\'"%message)
    except:
        print 'no commit'
    local("git push coding master")
    local("git checkout release-dev")
    local("git merge master --no-ff")
    local("git push coding release-dev")
    local("git checkout release-stable")
    local("git merge release-dev --no-ff")
    local("git push coding release-stable")
    local("git checkout master")

def reset():
    local("python radiusctl initdb -c etc/toughradius.json")
    local("python radiusctl inittest -c etc/toughradius.json")
    local("python radiusctl standalone -c etc/toughradius.json")

def all():
    os.environ['XDEBUG'] = 'true'
    local("python radiusctl standalone -c etc/toughradius.json")

def initdb():
    local("python radiusctl initdb -c etc/toughradius.json")
    local("python radiusctl inittest -c etc/toughradius.json")

def uplib():
    local("venv/bin/pip install -U --no-deps https://github.com/talkincode/toughlib/archive/master.zip")
    local("venv/bin/pip install -U --no-deps https://github.com/talkincode/txradius/archive/master.zip")

def startup():
    os.environ['XDEBUG'] = 'true'
    local("supervisord -n -c etc/toughradius_test.conf")




