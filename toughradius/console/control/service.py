#!/usr/bin/env python
# coding:utf-8
import os
import subprocess

from bottle import Bottle

from toughradius.console.base import *

__prefix__ = "/service"

app = Bottle()
app.config['__prefix__'] = __prefix__


##############################################################################
# basic
##############################################################################

class ToughError(Exception):
    def __init__(self, message):
        self.message = message


def run_command(command, raise_error_on_fail=False, shell=True, env=None):
    _result = dict(code=0)
    run_env = os.environ.copy()
    if env: run_env.update(env)
    log.msg("start exec")
    proc = subprocess.Popen(command, shell=shell,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                            env=run_env)
    stdout, stderr = proc.communicate('through stdin to stdout')
    result = proc.returncode, stdout, stderr
    if proc.returncode > 0 and raise_error_on_fail:
        error_string = "* Could not run command (return code= %s)\n" % proc.returncode
        error_string += "* Error was:\n%s\n" % (stderr.strip())
        error_string += "* Command was:\n%s\n" % command
        error_string += "* Output was:\n%s\n" % (stdout.strip())
        if proc.returncode == 127:  # File not found, lets print path
            path = os.getenv("PATH")
            error_string += "Check if y/our path is correct: %s" % path
        raise ToughError(error_string)
    else:
        return result


def warp_html(code, value):
    _value = value.replace("\n", "<br>")
    _value = _value.replace("RUNNING", "<strong><font color=green>RUNNING</font></strong>")
    _value = _value.replace("STARTING", "<strong><font color='#CC9900'>STARTING</font></strong>")
    _value = _value.replace("FATAL", "<strong><font color=red>FATAL</font></strong>")
    if code > 0:
        _value = '<font color="#CC0000">%s</font>' % _value
    return _value


def execute(cmd):
    try:
        log.msg("execute %s"%cmd)
        rcode, stdout, stderr = run_command(cmd, True)
        return dict(value=warp_html(rcode, (stdout or stderr)))
    except ToughError, err:
        return dict(value=warp_html(1, err.message))

@app.post('/upgrade', apply=auth_ctl)
def do_upgrade(render):
    release = request.params.get("release")
    cmd1 = "cd /opt/toughradius"
    cmd2 = "git checkout %s && git pull origin %s  " % (release, release)
    cmd3 = "supervisorctl restart radiusd"
    cmd4 = "supervisorctl restart admin"
    cmd5 = "supervisorctl restart customer"
    return execute("%s && %s && %s && %s && %s"%(cmd1,cmd2,cmd3,cmd4,cmd5))


@app.get('/initdb', apply=auth_ctl)
def do_upgrade(render):
    return execute("/opt/toughradius/toughctl --initdb")


@app.get('/radiusd/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart radiusd && supervisorctl status radiusd")


@app.get('/admin/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart admin && supervisorctl status admin")


@app.get('/customer/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart customer && supervisorctl status customer")


@app.get('/control/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart control &&  supervisorctl status control")


@app.get('/radiusd/status', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl status radiusd")


@app.get('/admin/status', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl status admin")


@app.get('/customer/status', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl status customer")


@app.get('/control/status', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl status control")