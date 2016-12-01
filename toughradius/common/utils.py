#!/usr/bin/env python
#coding:utf-8
import decimal
import datetime
from Crypto.Cipher import AES
from Crypto import Random
import hashlib
import binascii
import hashlib
import base64
import calendar
import random
import os
import time
import uuid
import json
import functools
import logging
import urlparse

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

    def is_pwd_encrypt(self):
        return os.environ.get("CLOSE_PASSWORD_ENCRYPTION")

    def setup(self, key): 
        self.bs = 32
        self.ori_key = key
        self.key = hashlib.sha256(key.encode()).digest()

    def encrypt(self, raw):
        is_encrypt = self.is_pwd_encrypt()
        if is_encrypt:
            return raw

        raw = safestr(raw)
        raw = self._pad(raw)
        iv = Random.new().read(AES.block_size)
        cipher = AES.new(self.key, AES.MODE_CBC, iv)
        return base64.b64encode(iv + cipher.encrypt(raw))

    def decrypt(self, enc):
        is_encrypt = self.is_pwd_encrypt()
        if is_encrypt:
            return enc
            
        enc = base64.b64decode(enc)
        iv = enc[:AES.block_size]
        cipher = AES.new(self.key, AES.MODE_CBC, iv)
        return safeunicode(self._unpad(cipher.decrypt(enc[AES.block_size:])))

    def _pad(self, s):
        return s + (self.bs - len(s) % self.bs) * chr(self.bs - len(s) % self.bs)

    def _unpad(self,s):
        return s[:-ord(s[len(s)-1:])]

aescipher = AESCipher()
encrypt = aescipher.encrypt
decrypt = aescipher.decrypt 

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
    
def kb2mb(ik,fmt='1.00'):
    _kb = decimal.Decimal(ik or 0)
    _mb = _kb / decimal.Decimal(1024)
    return str(_mb.quantize(decimal.Decimal(fmt)))
    
def mb2kb(im=0):
    _mb = decimal.Decimal(im or 0)
    _kb = _mb * decimal.Decimal(1024)
    return int(_kb.to_integral_value())    

def kb2gb(ik,fmt='1.00'):
    _kb = decimal.Decimal(ik or 0)
    _mb = _kb / decimal.Decimal(1024*1024)
    return str(_mb.quantize(decimal.Decimal(fmt)))
    
def gb2kb(im=0):
    _mb = decimal.Decimal(im or 0)
    _kb = _mb * decimal.Decimal(1024*1024)
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

def get_datetime(second):
    return time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(second))

def datetime2msec(dtime_str):
    _datetime =  datetime.datetime.strptime(dtime_str,"%Y-%m-%d %H:%M:%S")
    return int(time.mktime(_datetime.timetuple()))
    
def gen_backup_id():
    global _base_id
    if _base_id >= 9999:_base_id=0
    _base_id += 1
    _num = str(_base_id).zfill(4)
    return datetime.datetime.now().strftime("%Y%m%d_%H%M%S_") + _num

gen_backep_id = gen_backup_id

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
    try:
        expire_date = datetime.datetime.strptime("%s 23:59:59" % dstr, "%Y-%m-%d %H:%M:%S")
        now = datetime.datetime.now()
        return expire_date < now
    except:
        import traceback
        traceback.print_exc()
        return False

def fmt_online_time(ctime):
    if not ctime:
        return ''

    cdate = datetime.datetime.strptime(ctime, '%Y-%m-%d %H:%M:%S')
    nowdate = datetime.datetime.now()
    dt = nowdate - cdate
    times = dt.total_seconds()
    if times <= 60:
        return u"%s秒"%int(times)

    d = times / (3600 * 24)
    h = times % (3600 * 24) / 3600
    m = times % (3600 * 24) % 3600 / 60
    s = times % (3600 * 24) % 3600 % 60

    if int(d) > 0:
        return u"%s天%s小时%s分钟%s秒" % (int(d), int(h), int(m),int(s))
    elif int(d) == 0 and int(h) > 0:
        return u"%s小时%s分钟%s秒" % (int(h), int(m), int(s))
    elif int(d) == 0 and int(h) == 0 and int(m) > 0:
        return u"%s分钟%s秒" % (int(m),int(s))


def add_months(dt,months, days=0):
    month = dt.month - 1 + months
    year = dt.year + month / 12
    month = month % 12 + 1
    day = min(dt.day,calendar.monthrange(year,month)[1])
    dt = dt.replace(year=year, month=month, day=day)
    return dt + datetime.timedelta(days=days)


def is_connect(timestr, period=600):
    if not timestr:
        return False
    try:
        last_ping = datetime.datetime.strptime(timestr, "%Y-%m-%d %H:%M:%S")
        now = datetime.datetime.now()
        tt = now - last_ping
        return tt.seconds < period
    except:
        return False

def serial_model(mdl):
    if not mdl:return
    if not hasattr(mdl,'__table__'):return
    data = {}
    for c in mdl.__table__.columns:
        data[c.name] = getattr(mdl, c.name)
    return json.dumps(data,ensure_ascii=False)

def safestr(val):
    if val is None:
        return ''

    if isinstance(val, unicode):
        try:
            return val.encode('utf-8')
        except:
            return val.encode('gb2312')
    elif isinstance(val, str):
        return val
    elif isinstance(val, int):
        return str(val)
    elif isinstance(val, float):
        return str(val)
    elif isinstance(val, (dict,list)):
        return json.dumps(val, ensure_ascii=False)
    else:
        try:
            return str(val)
        except:
            return val
    return val

def safeunicode(val):
    if val is None:
        return u''

    if isinstance(val, str):
        try:
            return val.decode('utf-8')
        except:
            try:
                return val.decode('gb2312')
            except:
                return val
    elif isinstance(val, unicode):
        return val
    elif isinstance(val, int):
        return str(val).decode('utf-8')
    elif isinstance(val, float):
        return str(val).decode('utf-8')
    elif isinstance(val, (dict,list)):
        return json.dumps(val)
    else:
        try:
            return str(val).decode('utf-8')
        except:
            return val
    return val

def gen_secret(clen=32):
    rg = random.SystemRandom()
    r = list('1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ')
    return ''.join([rg.choice(r) for _ in range(clen)])

def timecast(func):
    from twisted.python import log
    @functools.wraps(func)
    def warp(*args,**kargs):
        _start = time.clock()
        result = func(*args,**kargs)
        log.msg("%s cast %.6f second"%(func.__name__,time.clock()-_start))
        return result
    return warp

def split_mline(src,wd=32,rstr='\r\n'):
    _idx = 0
    ss = []
    for c in src:
        if _idx > 0 and _idx%wd == 0:
            ss.append(rstr)
        ss.append(c)
        _idx += 1
    return ''.join(ss)


def get_cron_interval(cron_time):
    if cron_time:
        cron_time = "%s:00"%cron_time
        date_now = datetime.datetime.now()
        _now_hm = date_now.strftime("%H:%M:%S")
        _ymd = get_currdate()
        if _now_hm  > cron_time:
            _ymd = (date_now + datetime.timedelta(days=1)).strftime("%Y-%m-%d") 
        _interval = datetime.datetime.strptime("%s %s"%(_ymd,cron_time),"%Y-%m-%d %H:%M:%S") - date_now
        _itimes = int(_interval.total_seconds())
        return _itimes if _itimes > 0 else 86400 
    else:
        return 120



if __name__ == '__main__':
    aes = AESCipher("LpWE9AtfDPQ3ufXBS6gJ37WW8TnSF920")
    # aa = aes.encrypt(u"中文".encode('utf-8'))
    # print aa
    # cc = aes.decrypt(aa)
    # print cc.encode('utf-8')
    # aa = aes.decrypt("+//J9HPYQ+5PccoBZml6ngcLLu1/XQh2KyWakfcExJeb0wyq1C9+okztyaFbspYZ")
    # print aa
    # print get_cron_interval('09:32') 
    now = datetime.datetime.now()
    mon = now.month + 1
    mon = mon if mon <= 12 else 1
    timestr = "%s-%s-1 01:00:00" % (now.year,mon)
    _date = datetime.datetime.strptime(timestr, "%Y-%m-%d %H:%M:%S")
    tt = (time.mktime(_date.timetuple()) - time.time()) /86400
    print _date,tt









