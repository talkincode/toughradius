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
import time
import bottle
import decimal
import datetime
import functools
import subprocess
import platform


app = Bottle()


@app.route('/static/:path#.+#')
def route_static(path, render):
    static_path = os.path.join(os.path.split(os.path.split(__file__)[0])[0], 'static')
    return static_file(path, root=static_path)


@app.get('/', apply=auth_ctl)
def control_index(render):
    return render("index")


@app.route('/dashboard', apply=auth_ctl)
def index(render):
    return render("index", **locals())




