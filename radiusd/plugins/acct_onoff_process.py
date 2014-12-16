#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from store import store
from settings import *
import logging


"""记账启动关闭处理"""
def process(req=None,user=None,runstat=None):
    if  req.get_acct_status_type() not in (STATUS_TYPE_ACCT_ON,STATUS_TYPE_ACCT_OFF):
        return
        
    onlines = store.del_nas_onlines(req.get_nas_addr()) 

    if req.get_acct_status_type() == STATUS_TYPE_ACCT_ON:
        runstat.acct_on += 1  
        log.msg('bas accounting on success',level=logging.INFO)
    else:
        runstat.acct_off += 1  
        log.msg('bas accounting off success',level=logging.INFO)