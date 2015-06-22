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


__prefix__ = "/config"

app = Bottle()
app.config['__prefix__'] = __prefix__


@app.get('/', apply=auth_ctl)
def control_config(render):
    active = request.params.get("active","default")
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    default_form = config_forms.default_form()
    default_form.fill(dict(cfg.items("DEFAULT")))
    database_form = config_forms.database_form()
    database_form.fill(dict(cfg.items("database")))
    radiusd_form = config_forms.radiusd_form()
    radiusd_form.fill(dict(cfg.items("radiusd")))
    admin_form = config_forms.admin_form()
    admin_form.fill(dict(cfg.items("admin")))
    customer_form = config_forms.customer_form()
    customer_form.fill(dict(cfg.items("customer")))
    control_form = config_forms.control_form()
    control_form.fill(dict(cfg.items("control")))

    return render("config",
                  active=active,
                  default_form=default_form,
                  database_form=database_form,
                  radiusd_form=radiusd_form,
                  admin_form=admin_form,
                  customer_form=customer_form,
                  control_form=control_form
                  )



@app.post('/default/update', apply=auth_ctl)
def update_default(render):
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    cfg.set('DEFAULT','debug',request.forms.get("debug"))
    cfg.set('DEFAULT','tz',request.forms.get("tz"))
    cfg.set('DEFAULT','ssl',request.forms.get("ssl"))
    cfg.set('DEFAULT','privatekey',request.forms.get("privatekey"))
    cfg.set('DEFAULT','certificate',request.forms.get("certificate"))

    with open(app.config['DEFAULT.cfgfile'],'wb') as configfile:
        cfg.write(configfile)

    redirect("/config?active=default")


@app.post('/database/update', apply=auth_ctl)
def update_default(render):
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    cfg.set('database', 'echo', request.forms.get("echo"))
    cfg.set('database', 'dbtype', request.forms.get("dbtype"))
    cfg.set('database', 'dburl', request.forms.get("dburl"))
    cfg.set('database', 'pool_size', request.forms.get("pool_size"))
    cfg.set('database', 'pool_recycle', request.forms.get("pool_recycle"))
    cfg.set('database', 'backup_path', request.forms.get("backup_path"))

    with open(app.config['DEFAULT.cfgfile'], 'wb') as configfile:
        cfg.write(configfile)

    redirect("/config?active=database")


@app.post('/radiusd/update', apply=auth_ctl)
def update_default(render):
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    cfg.set('radiusd', 'host', request.forms.get("host"))
    cfg.set('radiusd', 'acctport', request.forms.get("acctport"))
    cfg.set('radiusd', 'adminport', request.forms.get("adminport"))
    cfg.set('radiusd', 'authport', request.forms.get("authport"))
    cfg.set('radiusd', 'cache_timeout', request.forms.get("cache_timeout"))
    cfg.set('radiusd', 'logfile', request.forms.get("logfile"))

    with open(app.config['DEFAULT.cfgfile'], 'wb') as configfile:
        cfg.write(configfile)

    redirect("/config?active=radiusd")


@app.post('/admin/update', apply=auth_ctl)
def update_default(render):
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    cfg.set('admin', 'host', request.forms.get("host"))
    cfg.set('admin', 'port', request.forms.get("port"))
    cfg.set('admin', 'logfile', request.forms.get("logfile"))

    with open(app.config['DEFAULT.cfgfile'], 'wb') as configfile:
        cfg.write(configfile)

    redirect("/config?active=admin")


@app.post('/customer/update', apply=auth_ctl)
def update_default(render):
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    cfg.set('customer', 'host', request.forms.get("host"))
    cfg.set('customer', 'port', request.forms.get("port"))
    cfg.set('customer', 'logfile', request.forms.get("logfile"))

    with open(app.config['DEFAULT.cfgfile'], 'wb') as configfile:
        cfg.write(configfile)

    redirect("/config?active=customer")


@app.post('/control/update', apply=auth_ctl)
def update_default(render):
    cfg = ConfigParser.ConfigParser()
    cfg.read(app.config['DEFAULT.cfgfile'])
    cfg.set('control', 'host', request.forms.get("host"))
    cfg.set('control', 'port', request.forms.get("port"))
    cfg.set('control', 'user', request.forms.get("user"))
    if request.forms.get("passwd"):
        cfg.set('control', 'passwd', request.forms.get("passwd"))
    cfg.set('control', 'logfile', request.forms.get("logfile"))

    with open(app.config['DEFAULT.cfgfile'], 'wb') as configfile:
        cfg.write(configfile)

    redirect("/config?active=control")