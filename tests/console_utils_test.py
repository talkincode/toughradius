#!/usr/bin/env python
#coding:utf-8
from console.libs import utils

def TestMb2kb():
    m = None
    kb = utils.mb2kb(m)
    assert kb == 0 
    k = 0
    mb = utils.kb2mb(k)
    assert mb == '0.00'
    