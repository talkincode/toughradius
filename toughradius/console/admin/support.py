#!/usr/bin/env python
# coding=utf-8

import sys, os
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import abort
from beaker.cache import cache_managers
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import forms
from hashlib import md5
from twisted.python import log
import bottle
import datetime
import json
import functools

__prefix__ = "/support"

app = Bottle()
app.config['__prefix__'] = __prefix__
render = functools.partial(Render.render_app, app)


@app.route('/', method=['GET', 'POST'])
def support(db):
    return render("support")