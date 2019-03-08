#!/usr/bin/env python
import sys,os,time
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
import datetime

__version__ = "v6.0.1"

currtime = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

def push():
    message = raw_input("commit msg:")
    local("git add .")
    try:
        local("git commit -m \"%s - %s: %s\""%(__version__, currtime, message))
    except:
        print 'no commit'
    local("git push origin develop")
    # local("git push coding master")
    # local("git push osc master")

def tag():
    local("git tag -a v%s -m 'version %s'"%(__version__,__version__))
    local("git push origin v%s:v%s"%(__version__,__version__))



def push_stable():
    message = raw_input("commit msg:")
    local("git add .")
    try:
        local("git commit -am \'%s\'"%message)
    except:
        print 'no commit'
    local("git push origin develop")
    local("git checkout master")
    local("git merge develop --no-ff")
    local("git push origin master")
    local("git checkout develop")




