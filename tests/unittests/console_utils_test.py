#!/usr/bin/env python
#coding:utf-8
from toughradius.console.libs import utils

def test_mb2kb2mb():
    assert utils.mb2kb(0) == 0 
    assert utils.mb2kb('') == 0 
    assert utils.mb2kb(None) == 0 
    assert utils.kb2mb(0) == '0.00'
    assert utils.kb2mb(None) == '0.00'
    assert utils.kb2mb('') == '0.00'
    