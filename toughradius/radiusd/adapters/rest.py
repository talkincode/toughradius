#!/usr/bin/env python
#coding:utf-8

from base import BasicAdapter
from toughradius.common import tools
from toughradius.common.mcache import cache
from hashlib import md5
import urllib2
import urllib
import json

class RestError(BaseException):pass

class RestAdapter(BasicAdapter):
    """Http rest mode adapter"""

    def getClient(self, nasip=None, nasid=None):
        def fetch_result():
            url = self.settings.ADAPTERS['rest']['nasurl']
            appid=self.settings.ADAPTERS['rest']['appid']
            secret=self.settings.ADAPTERS['rest']['secret']
            msg = dict(
                appid=appid,
                nasid=nasid or '',
                nasip=nasip or ''
            )
            msg['sign'] = self.makeSign(secret, msg.values())
            try:
                req = urllib2.Request(url, urllib.urlencode(msg) )
                resp = urllib2.urlopen(req,timeout=5)
                result = json.loads(resp.read())
                if result['code'] > 0:
                    raise RestError("rest request error %s" % result.get('msg',''))
                else:
                    return result['data']
            except:
                raise RestError("rest request error")
        return cache.aget('toughradius.nas.cache.{0}.{1}'.format(nasid,nasip),fetch_result,expire=60)


    def makeSign(self,api_secret,params):
        _params = [tools.safeunicode(p) for p in params if p is not None]
        _params.sort()
        # print 'sorted params:',_params
        _params.insert(0, api_secret)
        strs = ''.join(_params)
        # print 'sign params:',strs
        mds = md5(strs.encode('utf-8')).hexdigest()
        return mds.upper()


    def processAuth(self,req):
        url = self.settings.ADAPTERS['rest']['authurl']
        appid = self.settings.ADAPTERS['rest']['appid']
        secret = self.settings.ADAPTERS['rest']['secret']
        msg = req.dict_message
        msg['appid'] = appid
        msg['sign'] = self.makeSign(secret, msg.values())
        try:
            req = urllib2.Request(url, urllib.urlencode(msg))
            resp = urllib2.urlopen(req,timeout=5)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")


    def processAcct(self,req):
        url = self.settings.ADAPTERS['rest']['accturl']
        appid = self.settings.ADAPTERS['rest']['appid']
        secret = self.settings.ADAPTERS['rest']['secret']
        msg = req.dict_message
        msg['appid'] = appid
        msg['nas_paddr'] = req.source[0]
        msg['sign'] = self.makeSign(secret, msg.values())
        try:
            req = urllib2.Request(url, urllib.urlencode(msg))
            resp = urllib2.urlopen(req,timeout=5)
            return json.loads(resp.read())
        except:
            raise RestError("rest request error")


adapter = RestAdapter