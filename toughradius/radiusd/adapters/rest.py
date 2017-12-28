#!/usr/bin/env python
#coding:utf-8

from base import BasicAdapter
from toughradius.common import tools
from toughradius.common.mcache import cache
from hashlib import md5
import urllib2
import urllib
import json
import logging

logger = logging.getLogger(__name__)

class RestError(BaseException):pass

class RestAdapter(BasicAdapter):
    """Http rest mode adapter"""

    @tools.timecast
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
                    logger.error("rest request error %s" % result.get('msg',''))
                    return None
                else:
                    return result['data']
            except:
                logger.error("rest request error", exc_info=True)
                return None
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

    @tools.timecast
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
        except Exception as err:
            logger.error("radius rest auth error", exc_info=True)
            return dict(code=1, msg=err.message)

    @tools.timecast
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
        except Exception as err:
            logger.error("radius rest acct error", exc_info=True)
            return dict(code=1, msg=err.message)


adapter = RestAdapter