#!/usr/bin/env python
#coding:utf-8
from twisted.web import server, wsgi
from twisted.python.threadpool import ThreadPool
from twisted.internet import reactor
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import abort
import functools
import bottle
import datetime
import json

app = Bottle()

def doauth(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        return func(*args,**kargs)
    return warp

@app.route('/user/auth',apply=doauth)
def user_auth():    
    return dict(code=0,msg="hello")

@app.route('/user/reauth',apply=doauth)
def user_reauth():    
    return dict(code=0,msg="hello")

@app.route('/user/unauth',apply=doauth)
def user_unauth():    
    return dict(code=0,msg="hello")

@app.error(404)
def error404(error):
    return dict(code=404,msg='error 404')

@app.error(500)
def error500(error):
    return dict(code=404,msg='error 500 %s'%error.exception)

def setup(port):
    if not port:return
    thread_pool = ThreadPool()
    thread_pool.start()
    reactor.addSystemEventTrigger('after', 'shutdown', thread_pool.stop)
    factory = server.Site(wsgi.WSGIResource(reactor, thread_pool, app))
    reactor.listenTCP(port, factory, interface='0.0.0.0')







