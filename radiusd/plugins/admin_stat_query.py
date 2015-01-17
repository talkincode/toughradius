#!/usr/bin/env python
#coding=utf-8
import json

def process(req=None,admin=None):
    admin.sendMessage(json.dumps(admin.runstat),False) 


