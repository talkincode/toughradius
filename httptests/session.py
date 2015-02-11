#!/usr/bin/env python
#coding:utf-8
import requests

url = "http://127.0.0.1:1816"
login_url = "%s/login"%url

def sub_path(path):
    return "%s%s"%(url,path)

def login():
    req = requests.Session()
    r = req.post(login_url,data=dict(username="admin",password="root"))
    if r.status_code == 200:
        rjson =  r.json()
        msg = rjson['msg']
        if rjson['code'] == 0:
            return req
        else:
            raise Exception(msg)
    else:
        r.raise_for_status()

if __name__ == '__main__':
    login()