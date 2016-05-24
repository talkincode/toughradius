#!/usr/bin/env python
import sys,os,time,datetime
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
from toughradius import __version__

# --------------------------- remote deploy--------------------------

env.user = 'root'
env.hosts = ['121.201.15.99']

def push():
    message = raw_input("commit msg:")
    local("git add .")
    try:
        local("git commit -m \'%s:%s\'"%(__version__,message))
    except:
        print 'no commit'
    local("git push origin master")
    local("git push osc master")
    
def deploy():
    gitrepo = "git@github.com:talkincode/ToughRADIUS.git"
    rundir = "/opt/toughradius"
    run("test -d {rundir} || git clone -b master {gitrepo} {rundir}".format(rundir=rundir,gitrepo=gitrepo))
    with cd(rundir):
        run("git pull --rebase --stat origin master")
        run("make all")
        run("make initdb")
        run("service toughradius restart")
        run("service toughradius status")

def upgrade():
    with cd("/opt/toughradius"):
        run("git pull --rebase --stat origin master")
        run("service toughradius restart")
        run("service toughradius status")

def restart():
    run("service toughradius restart")
    run("service toughradius status")

def tail():
    run("tail -f /var/toughradius/radius-manage.log")

def tail100():
    run("tail -n 100 /var/toughradius/radius-manage.log")

def status():
    run("service toughradius status")

def uplib():
    with cd("/opt/toughradius"):
        run("make upgrade-libs")


# -----------------------------------------------------

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

def push_dev():
    message = raw_input("commit msg:")
    local("git add .")
    local("git commit -m \'%s\'"%message)
    local("git push origin master")
    local("git checkout release-dev")
    local("git pull origin release-dev")
    local("git merge master --no-ff")
    local("git push origin release-dev")
    local("git checkout master")

def push_stable():
    message = raw_input("commit msg:")
    local("git add .")
    local("git commit -m \'%s\'"%message)
    local("git push origin master")
    local("git checkout release-dev")
    local("git merge master --no-ff")
    local("git push origin release-dev")
    local("git checkout release-stable")
    local("git merge release-dev --no-ff")
    local("git push origin release-stable")
    local("git checkout master")

def all():
    local("venv/bin/python radiusctl standalone -c ~/toughradius_test.json")


def reset():
    local("venv/bin/python radiusctl initdb -c ~/toughradius_test.json")
    local("venv/bin/python radiusctl inittest -c ~/toughradius_test.json")  
    local("venv/bin/python radiusctl standalone -c ~/toughradius_test.json")



def initdb():
    local("venv/bin/python radiusctl initdb -c ~/toughradius_test.json")
    local("venv/bin/python radiusctl inittest -c ~/toughradius_test.json")

