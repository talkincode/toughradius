#!/usr/bin/env python
#coding:utf-8
import decimal
import datetime
from Crypto.Cipher import AES
from Crypto import Random
import json
import hashlib
import base64
import calendar
import random
import os
import time
import uuid

random_generator = random.SystemRandom()

decimal.getcontext().prec = 32
decimal.getcontext().rounding = decimal.ROUND_UP

_base_id = 0

_CurrentID = random_generator.randrange(1, 1024)

def CurrentID():
    global _CurrentID
    _CurrentID = (_CurrentID + 1) % 1024
    return str(_CurrentID)

class AESCipher:
    
    def __init__(self,key=None):
        if key:self.setup(key)

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
        return s[:-ord(s[len(s)-1:])]

aescipher = AESCipher()
encrypt = aescipher.encrypt
decrypt = aescipher.decrypt 

def mk_sign(args=[]):
    args.sort()
    _argstr =  aescipher.ori_key + ''.join(args)
    return hashlib.md5(_argstr).hexdigest()

def update_tz(tz_val,default_val="CST-8"):
    try:
        os.environ["TZ"] = tz_val or default_val
        time.tzset()
    except:
        pass
        
def check_ssl(config):
    use_ssl = False
    privatekey = None
    certificate = None
    if config.has_option('DEFAULT','ssl') and config.getboolean('DEFAULT','ssl'):
        privatekey = config.get('DEFAULT','privatekey')
        certificate = config.get('DEFAULT','certificate')
        if os.path.exists(privatekey) and os.path.exists(certificate):
            use_ssl = True
    return use_ssl,privatekey,certificate
    
def get_uuid():
    return uuid.uuid1().hex.upper()
    
def bps2mbps(bps):
    _bps = decimal.Decimal(bps or 0)
    _mbps = _bps / decimal.Decimal(1024*1024)
    return str(_mbps.quantize(decimal.Decimal('1.000')))
    
def mbps2bps(mbps):
    _mbps = decimal.Decimal(mbps or 0)
    _kbps = _mbps * decimal.Decimal(1024*1024)
    return int(_kbps.to_integral_value())
    
def bb2mb(ik):
    _kb = decimal.Decimal(ik or 0)
    _mb = _kb / decimal.Decimal(1024*1024)
    return str(_mb.quantize(decimal.Decimal('1.00')))
    
def bbgb2mb(bb,gb):
    bl = decimal.Decimal(bb or 0)/decimal.Decimal(1024*1024)
    gl = decimal.Decimal(gb or 0)*decimal.Decimal(4*1024*1024*1024)
    tl = bl + gl
    return str(tl.quantize(decimal.Decimal('1.00')))
    
def kb2mb(ik):
    _kb = decimal.Decimal(ik or 0)
    _mb = _kb / decimal.Decimal(1024)
    return str(_mb.quantize(decimal.Decimal('1.00')))
    
def mb2kb(im=0):
    _mb = decimal.Decimal(im or 0)
    _kb = _mb * decimal.Decimal(1024)
    return int(_kb.to_integral_value())
    
def hour2sec(hor=0):
    _hor = decimal.Decimal(hor or 0)
    _sec = _hor * decimal.Decimal(3600)
    return int(_sec.to_integral_value())

def sec2hour(sec=0):
    _sec = decimal.Decimal(sec or 0)
    _hor = _sec / decimal.Decimal(3600)
    return str(_hor.quantize(decimal.Decimal('1.00')))

def fen2yuan(fen=0):
    f = decimal.Decimal(fen or 0)
    y = f / decimal.Decimal(100)
    return str(y.quantize(decimal.Decimal('1.00')))

def yuan2fen(yuan=0):
    y = decimal.Decimal(yuan or 0)
    f = y * decimal.Decimal(100)
    return int(f.to_integral_value())

def get_currtime():
    return datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

def get_currdate():
    return datetime.datetime.now().strftime("%Y-%m-%d") 
    
def gen_backep_id():
    global _base_id
    if _base_id >= 9999:_base_id=0
    _base_id += 1
    _num = str(_base_id).zfill(4)
    return datetime.datetime.now().strftime("%Y%m%d_%H%M%S_") + _num

def gen_order_id():
    global _base_id
    if _base_id >= 9999:_base_id=0
    _base_id += 1
    _num = str(_base_id).zfill(4)
    return datetime.datetime.now().strftime("%Y%m%d%H%M%S") + _num

def fmt_second(time_total):
    """
    >>> fmt_second(100)
    '00:01:40'
    """

    def _ck(t):
        return t < 10 and "0%s" % t or t

    times = int(time_total)
    h = times / 3600
    m = times % 3600 / 60
    s = times % 3600 % 60
    return "%s:%s:%s" % (_ck(h), _ck(m), _ck(s))


def is_expire(dstr):
    if not dstr:
        return False
    expire_date = datetime.datetime.strptime("%s 23:59:59" % dstr, "%Y-%m-%d %H:%M:%S")
    now = datetime.datetime.now()
    return expire_date < now

def fmt_online_time(ctime):
    if not ctime:
        return ''

    cdate = datetime.datetime.strptime(ctime, '%Y-%m-%d %H:%M:%S')
    nowdate = datetime.datetime.now()
    dt = nowdate - cdate
    times = dt.total_seconds()
    d = times / (3600 * 24)
    h = times % (3600 * 24) / 3600
    m = times % (3600 * 24) % 3600 / 60

    if int(d) > 0:
        return u"%s天%s小时%s分钟" % (int(d), int(h), int(m))
    elif int(d) > 0 and int(h) > 0:
        return u"%s小时%s分钟" % (int(h), int(m))
    else:
        return u"%s分钟" % (int(m))


def add_months(dt,months):
    month = dt.month - 1 + months
    year = dt.year + month / 12
    month = month % 12 + 1
    day = min(dt.day,calendar.monthrange(year,month)[1])
    return dt.replace(year=year, month=month, day=day)

def safestr(val):
    if val is None:
        return ''
    elif isinstance(val, unicode):
        return val.encode('utf-8')
    elif isinstance(val, str):
        return val
    elif isinstance(val, int):
        return str(val)
    elif isinstance(val, float):
        return str(val)
    elif isinstance(val, dict):
        return json.dumps(val, ensure_ascii=False)
    return val

if __name__ == '__main__':
    print gen_order_id()
    print gen_order_id()
