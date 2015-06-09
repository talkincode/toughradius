#!/usr/bin/env python
# coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import forms
from sqlalchemy import func
import bottle
import datetime

__prefix__ = "/user_stat"

app = Bottle()
app.config['__prefix__'] = __prefix__

