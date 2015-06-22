#!/usr/bin/env python
# coding:utf-8
import sys, os
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
from toughradius.console.libs.validate import vcache
from toughradius.console.control import config_forms
import time
import bottle
import decimal
import datetime
import functools
import ConfigParser

__prefix__ = "/logquery"

app = Bottle()
app.config['__prefix__'] = __prefix__

@app.get('/:name', apply=auth_ctl)
def logquery(name, render):
    def _query(logfile):
        if os.path.exists(logfile):
            with open(logfile) as f:
                f.seek(0, 2)
                if f.tell() > 32 * 1024:
                    f.seek(f.tell() - 32 * 1024)
                else:
                    f.seek(0)
                return f.read().replace('\n', '<br>')

    if '%s.logfile' % name in app.config:
        logfile = app.config['%s.logfile' % name]
        return render("logquery", msg=_query(logfile), title="%s logging" % name, logfile=logfile)
    else:
        return render("logquery", msg="logfile not exists", title="%s logging" % name, logfile="logfile")
