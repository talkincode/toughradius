#!/usr/bin/env python
import sys,os,time,datetime
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
from toughradius import __version__

env.user = 'root'
env.hosts = ['121.201.63.77']

def build():
    releases = {'test':'master','dev':'release-dev','stable':'release-stable'}
    release = releases.get(raw_input("Please enter release type [test,dev,stable](default:dev):"),'dev')
    build_ver = "linux-{0}-{1}".format(release, datetime.datetime.now().strftime("%Y%m%d%H%M%S"))
    gitrepo = "https://github.com/talkincode/ToughRADIUS.git"
    rundir = "/opt/toughradius"
    dist = "toughradius-{0}.tar.bz2".format(build_ver)
    run("test -d {0} || git clone {1} {2}".format(rundir,gitrepo,rundir))
    with cd(rundir):
        run("git checkout {0} && git pull -f origin {0}".format(release,release))
        run("make venv")
    with cd("/opt"):
        _excludes = ['.git','fabfile.py','pymodules','.travis.yml','.gitignore','dist',
        'coverage.txt','.coverage','.coverageerc','build','_trial_temp']
        excludes = ' '.join( '--exclude %s'%_e for _e in _excludes )
        run("tar -jpcv -f /tmp/{0} toughradius {1}".format(dist,excludes))
    local("scp  root@121.201.63.77:/tmp/{0} {1}".format(dist,dist))

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
    local("git merge master --no-ff")
    local("git push origin release-dev")
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

