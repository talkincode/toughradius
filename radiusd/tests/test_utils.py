#!/usr/bin/python
#coding:utf-8
import test_config
from radiusd import utils

def test_encrypt_decrypt():
    a = utils.encrypt('888888')
    assert a
    b = utils.decrypt(a)
    assert b == '888888'

def test_is_expire():
    assert  not utils.is_expire('')
    assert  not utils.is_expire('3000-12-30')
    assert  utils.is_expire('2013-12-10')


def test_delay():
    d = utils.AuthDelay(6)
    assert d.delay_len() == 0
    reject = {'id':1}
    d.add_delay_reject(reject)
    assert d.delay_len() == 1
    assert d.pop_delay_reject() == reject 
    assert d.delay_len() == 0
    mac_addr = '12:1s:4d:2s:2d:s2'
    for i in range(7):
        d.add_roster(mac_addr)
    assert d.rosters[mac_addr] > 6
    assert d.over_reject(mac_addr)
