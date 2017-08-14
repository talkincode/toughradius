#!/usr/bin/env python
#coding:utf-8

from bottle import route, run, request,post

@post('/api/v1/radtest')
def radtest():
    return dict(code=0,msg="success")


def start(host='0.0.0.0',port=1815):
    run(server="gevent", host=host, port=port)


if __name__ == '__main__':
    import gevent.monkey
    gevent.monkey.patch_all()
    start()