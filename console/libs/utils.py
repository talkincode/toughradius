#!/usr/bin/env python
#coding:utf-8
import decimal
import datetime
from Crypto.Cipher import AES
from Crypto import Random
import binascii
import hashlib
import base64

decimal.getcontext().prec = 11
decimal.getcontext().rounding = decimal.ROUND_UP

_base_id = 0

_key = 't_o_u_g_h_radius'

class AESCipher:

    def __init__(self, key): 
        self.bs = 32
        self.key = hashlib.sha256(key.encode()).digest()

    def encrypt(self, raw):
        raw = self._pad(raw)
        iv = Random.new().read(AES.block_size)
        cipher = AES.new(self.key, AES.MODE_CBC, iv)
        return base64.b64encode(iv + cipher.encrypt(raw))

    def decrypt(self, enc):
        enc = base64.b64decode(enc)
        iv = enc[:AES.block_size]
        cipher = AES.new(self.key, AES.MODE_CBC, iv)
        return self._unpad(cipher.decrypt(enc[AES.block_size:])).decode('utf-8')

    def _pad(self, s):
        return s + (self.bs - len(s) % self.bs) * chr(self.bs - len(s) % self.bs)

    @staticmethod
    def _unpad(s):
        return s[:-ord(s[len(s)-1:])]

_aes = AESCipher(_key)
encrypt = _aes.encrypt
decrypt = _aes.decrypt 

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
