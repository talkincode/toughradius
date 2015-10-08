#!/usr/bin/env python
# coding:utf-8
import os

from bottle import Bottle

from toughradius.console.base import *

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




