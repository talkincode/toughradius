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

def uplog():
    push()
    upgrade()
    tail()


# def build():
#     releases = {'test':'master','dev':'release-dev','stable':'release-stable'}
#     release = releases.get(raw_input("Please enter release type [test,dev,stable](default:dev):"),'dev')
#     build_ver = "linux-{0}-{1}".format(release, datetime.datetime.now().strftime("%Y%m%d%H%M%S"))
#     gitrepo = "git@bitbucket.org:talkincode/toughradius-enterprise.git"
#     rundir = "/opt/toughradius"
#     dist = "toughradius-{0}.tar.bz2".format(build_ver)
#     run("test -d {0} || git clone {1} {2}".format(rundir,gitrepo,rundir))
#     with cd(rundir):
#         run("git checkout {0} && git pull -f origin {0}".format(release,release))
#         run("make venv")
#     with cd("/opt"):
#         _excludes = ['.git','fabfile.py','pymodules','.travis.yml','.gitignore','dist',
#         'coverage.txt','.coverage','.coverageerc','build','_trial_temp']
#         excludes = ' '.join( '--exclude %s'%_e for _e in _excludes )
#         run("tar -jpcv -f /tmp/{0} toughradius {1}".format(dist,excludes))
#     local("scp  root@121.201.63.77:/tmp/{0} {1}".format(dist,dist))

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

def sspd():
    os.environ['XDEBUG'] = 'true'
    local("python radiusctl ssportal -c etc/toughradius.json")

def initdb():
    local("python radiusctl initdb -c etc/toughradius.json")
    local("python radiusctl inittest -c etc/toughradius.json")

def uplib():
    local("venv/bin/pip install -U --no-deps https://github.com/talkincode/toughlib/archive/master.zip")
    # local("venv/bin/pip install -U --no-deps https://github.com/talkincode/txradius/archive/master.zip")





