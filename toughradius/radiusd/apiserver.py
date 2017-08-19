#!/usr/bin/env python
#coding:utf-8
import gevent.monkey
gevent.monkey.patch_all()
from toughradius.common.bottle import route, run, request,post
from toughradius.common.ghttpd import GeventServer

@post('/api/v1/radtest')
def radtest():
    return dict(code=0,msg="success")


def start(host='0.0.0.0',port=1815,forever=True):
    return run(server=GeventServer, host=host, port=port,forever=forever)


if __name__ == '__main__':
    start(forever=True)