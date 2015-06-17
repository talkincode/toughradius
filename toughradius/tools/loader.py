#!/usr/bin/env python
# coding:utf-8
from Crypto.Cipher import AES
from Crypto import Random
import hashlib
import binascii
import hashlib
import base64
import random
import uuid


class AESCipher:
    def __init__(self, key=None):
        if key: self.setup(key)

    def setup(self, key):
        self.bs = 32
        self.ori_key = key
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
        return s[:-ord(s[len(s) - 1:])]


aescipher = AESCipher(key="vC0JnIa7u5TzB6yYEXgWBqq63wBeNrGBAQuJmQLXkkRVGhtR6awIMkO8HmGQGrRa5ZvXMIHoFlwZ5mMtVgpEQtxCnoYal9UCDCsGZE3oT7VDPgepoLp1DscrLs6lgmet")


def check_lic(lic):
    def get_devid():
        node = uuid.getnode()
        mac = uuid.UUID(int=node).hex[-12:]
        return mac
    dlic = aescipher.decrypt(lic)
    return dlic == get_devid()

if __name__ == '__main__':
    print gen_secret(128)
    aa= encrypt(get_mac_address())
    print aa
