#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from toughradius.radiusd.settings import *
import logging


def process(req=None,user=None,radiusd=None,**kwargs):
    if  req.get_acct_status_type() not in (STATUS_TYPE_ACCT_ON,STATUS_TYPE_ACCT_OFF):
        return
    
    runstat=radiusd.runstat
    store = radiusd.store

    if req.get_acct_status_type() == STATUS_TYPE_ACCT_ON:
        store.unlock_online(req.get_nas_addr(),None,STATUS_TYPE_ACCT_ON)
        runstat.acct_on += 1  
        log.msg('bas accounting on success',level=logging.INFO)
    else:
        store.unlock_online(req.get_nas_addr(),None,STATUS_TYPE_ACCT_OFF)
        runstat.acct_off += 1  
        log.msg('bas accounting off success',level=logging.INFO)