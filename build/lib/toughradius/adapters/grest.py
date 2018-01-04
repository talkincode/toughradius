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
import os

_geventhttpclient = False
try:
    from geventhttpclient.client import HTTPClientPool
    from geventhttpclient.url import URL
    _geventhttpclient = True
except:
    import traceback
    traceback.print_exc()


logger = logging.getLogger(__name__)
trace = logging.getLogger("trace")


class RestError(BaseException):pass

class RestAdapter(BasicAdapter):
    """Http rest mode adapter"""

    def __init__(self, settings):
        BasicAdapter.__init__(self, settings)
        self.timeout = int(self.settings.ADAPTERS['rest'].get('timeout', 10))
        self.concurrency = int(self.settings.ADAPTERS['rest'].get('concurrency', 100))
        if _geventhttpclient:
            self.http_pool = HTTPClientPool(concurrency=self.concurrency, network_timeout=self.timeout)

    def request(self, url, msg):
        if os.environ.get('TOUGHRADIUS_TRACE_ENABLED', "0") == "1":
            trace.info("Send rest request: %s" % msg)
        if _geventhttpclient:
            url = URL(url)
            http = self.http_pool.get_client(url)
            resp = http.get(url.request_uri + "?" + urllib.urlencode(msg))
            resp_body = resp.read()
            if os.environ.get('TOUGHRADIUS_TRACE_ENABLED', "0") == "1":
                trace.info("Received rest response: %s" % resp_body)
            if resp.status_code == 200:
                return json.loads(resp_body)
        else:
            req = urllib2.Request(url, urllib.urlencode(msg))
            resp = urllib2.urlopen(req, timeout=self.timeout)
            resp_body = resp.read()
            if os.environ.get('TOUGHRADIUS_TRACE_ENABLED', "0") == "1":
                trace.info("Received rest response: %s" % resp_body)
            return json.loads(resp_body)


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
                result = self.request(url, msg)
                if result:
                    if result['code'] > 0:
                        logger.error("rest request error %s" % result.get('msg',''))
                    else:
                        return result['data']
            except:
                logger.error("rest request error", exc_info=True)
        return cache.aget('toughradius.nas.cache.{0}.{1}'.format(nasid,nasip),fetch_result,expire=30)


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
        msg = {k: ('' if v is None else v) for k,v in msg.iteritems()}
        msg['sign'] = self.makeSign(secret, msg.values())
        try:
            return self.request(url, msg)
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
        msg = {k: ('' if v is None else v) for k, v in msg.iteritems()}
        msg['sign'] = self.makeSign(secret, msg.values())
        try:
            return self.request(url, msg)
        except Exception as err:
            logger.error("radius rest acct error", exc_info=True)
            return dict(code=1, msg=err.message)


adapter = RestAdapter