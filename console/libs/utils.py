#!/usr/bin/env python
#coding:utf-8
import decimal
import datetime

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

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