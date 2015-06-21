#!/usr/bin/env python
# coding:utf-8
import sys, os
from twisted.python import log
from twisted.internet import reactor
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import abort
from hashlib import md5
from urlparse import urljoin
from toughradius.console.base import *
from toughradius.console.libs import utils
import time
import bottle
import decimal
import datetime
import functools
import subprocess
import platform

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

@app.get('/upgrade', apply=auth_ctl)
def do_upgrade(render):
    return execute("/usr/bin/toughrad upgrade")


@app.get('/radiusd/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart radiusd")


@app.get('/admin/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart admin")


@app.get('/customer/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart customer")


@app.get('/control/restart', apply=auth_ctl)
def do_upgrade(render):
    return execute("supervisorctl restart cuntrol")