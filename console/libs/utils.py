#!/usr/bin/env python
#coding:utf-8
import decimal
import datetime
from Crypto.Cipher import AES
import binascii

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

_base_id = 0

_key = 't_o_u_g_h_radius'


def encrypt(x):
    if not x:return ''
    x = str(x)
    result =  AES.new(_key, AES.MODE_CBC).encrypt(x.ljust(len(x)+(16-len(x)%16)))
    return binascii.hexlify(result)

def decrypt(x):
    if not x or len(x)%16 > 0 :return ''
    x = binascii.unhexlify(str(x))
    return AES.new(_key, AES.MODE_CBC).decrypt(x).strip()    

def fen2yuan(fen):
    f = decimal.Decimal(fen)
    y = f / decimal.Decimal(100)
    return str(y.quantize(decimal.Decimal('1.00')))

def yuan2fen(yuan):
    y = decimal.Decimal(yuan)
    f = y * decimal.Decimal(100)
    return int(f.to_integral_value())

def get_currtime():
    return datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

def get_currdate():
    return datetime.datetime.now().strftime("%Y-%m-%d") 

def gen_order_id():
    global _base_id
    if _base_id >= 9999:_base_id=0
    _base_id += 1
    _num = str(_base_id).zfill(4)
    return datetime.datetime.now().strftime("%Y%m%d%H%M%S") + _num


if __name__ == '__main__':
    print gen_order_id()
    print gen_order_id()
