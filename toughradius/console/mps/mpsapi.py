#coding:utf-8
import logging
import requests
import json
import urllib
import time
from cyclone import httpclient
from twisted.internet import defer
from toughradius.console.libs.apiutils import ApiMessage


class MpsApi:

    def __init__(self,config):
        # self.api_address = "http://mpsapi.lingyatech.net"
        self.api_address = "https://api.weixin.qq.com"
        self.oauth_address = "https://open.weixin.qq.com"
        self.upload_address = "http://file.api.weixin.qq.com"

    def wx_oauth_token_url(self,code):
        _url = "%s/sns/oauth2/access_token?appid=%s"\
        "&secret=%s&code=%s&grant_type=authorization_code"
        return _url % (
            self.api_address,
            config.get("mps","mps_appid"),
            config.get("mps","mps_secret"),
            code
        )

    def wx_oauth_redirect_url(self,oauthbak_url,**params):
        if 'http' not in oauthbak_url:
            oauthbak_url = "%s%s"%(config.get("mps","server_base"),oauthbak_url)
        _url = "%s/connect/oauth2/authorize?"\
        "appid=%s&redirect_uri=%s&response_type=code"\
        "&scope=snsapi_base&state=STATE#wechat_redirect"
        _params = params and "?"+urllib.urlencode(params) or ''
        return _url % (
            self.oauth_address,
            config.get("mps","mps_appid"),
            urllib.quote(oauthbak_url+_params)
        )

    def wx_gettoken_url(self):
        return '%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s' % \
               (self.api_address,config.get("mps","mps_appid"), config.get("mps","mps_secret"))     

    def wx_userinfo_url(self,openid):
        return '%s/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN'% \
        (self.api_address,self.get_mps_token(), openid)

    def wx_send_custommsg_url(self):
        return '%s/cgi-bin/message/custom/send?access_token=%s' % \
        (self.api_address,self.get_mps_token())

    def wx_sync_menus_url(self):
        return '%s/cgi-bin/menu/create?access_token=%s' % \
        (self.api_address,self.get_mps_token())

    def wx_sync_user_url(self):
        return '%s/cgi-bin/user/get?access_token=%s'%\
                (self.api_address,self.get_mps_token())

    def wx_upload_media_url(self,type):
        return '%s/cgi-bin/media/upload?access_token=%s&type=%s'%\
                (self.upload_address,self.get_mps_token(),type)

    def wx_upload_news_url(self):
        return '%s/cgi-bin/media/uploadnews?access_token=%s'%\
                (self.api_address,self.get_mps_token())

    def wx_push_msg_url(self):
        return '%s/cgi-bin/message/mass/send?access_token=%s'%\
                (self.api_address,self.get_mps_token())
                
    def wx_create_qrcode(self):
        return '%s/cgi-bin/qrcode/create?access_token=%s'%\
                (self.api_address,self.get_mps_token())

    def get_mps_token(self):
        _url = self.wx_gettoken_url()
        mps_access_token = config.get('mps','mps_access_token')
        if not mps_access_token:
            _resp = requests.get(_url)
            _json_obj = _resp.json()
            mps_access_token = _json_obj.get('access_token')
            mps_access_token_expires = time.time() + 6000
            if mps_access_token:
                logging.info('get a new access_token: ' + mps_access_token)
                config.set('mps','mps_access_token',mps_access_token)
                config.set('mps','mps_access_token_expires',mps_access_token_expires)
        else:
            if "mps_access_token_expires" not in config \
                    or config.getint('mps','mps_access_token_expires') < time.time():
                config.set('mps','mps_access_token_expires',0)
                return self.get_mps_token()

        return mps_access_token    

    def get_oauth_token(self,code):
        wx_url = self.wx_oauth_token_url(code) 
        wx_resp = requests.get(wx_url)
        return json.loads(wx_resp.text)
 
    def get_weixin_user(self,openid):
        wxu_url = self.wx_userinfo_url(openid)
        wxu_resp = requests.get(wxu_url)
        return json.loads(wxu_resp.text)

    def send_customer_text_msg(self,openid,msg):
        wxu_url = self.wx_send_custommsg_url()
        logging.info(wxu_url)
        _msg = dict(touser=openid, msgtype='text', text={'content': msg})
        _msg = json.dumps(_msg, ensure_ascii=False)
        return requests.post(wxu_url,_msg.encode("utf-8"))
        
    @defer.inlineCallbacks
    def send_customer_text_async(self,openid,msg):
        _url = self.wx_send_custommsg_url()
        _msg = dict(touser=openid, msgtype='text', text={'content': msg})
        _msg = json.dumps(_msg, ensure_ascii=False)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def send_customer_news_async(self,openid,title,description,url='',pic_url=''):
        _url = self.wx_send_custommsg_url()
        _msg = dict(touser=openid, msgtype='news', 
            news={'articles': [{"title":title,"description":description,"url":url, "picurl":pic_url}]})
        _msg = json.dumps(_msg, ensure_ascii=False)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def get_wx_user_info(self,openid):
        _url = self.wx_userinfo_url(openid)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(headers=_headers, use_gzip=True,connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def update_media(self,file,type):
        _url = self.wx_upload_media_url(type)
        files = {'media': ('update.jpg', open(file, 'rb'))}
        resp = requests.post(_url, files=files)
        if 'errcode' in resp.json() > 0:
            defer.returnValue(ApiMessage(code=1, msg=resp.json().get('errmsg')))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=resp.json()
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def update_news(self,media_id,title,content):
        _url = self.wx_upload_news_url()
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _msg = dict({'articles': [{"thumb_media_id":media_id,"author":None,"title":title, "content_source_url":None,\
                                   "content":content,"digest":content,"show_cover_pic":"1"}]})
        _msg = json.dumps(_msg, ensure_ascii=False)
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def push_news_msg(self,openid_list,media_id):
        _url = self.wx_push_msg_url()
        _msg = dict(touser=openid_list, msgtype='mpnews', mpnews={"media_id":media_id})
        _msg = json.dumps(_msg, ensure_ascii=False)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def push_text_msg(self,openid_list,content):
        _url = self.wx_push_msg_url()
        _msg = dict(touser=openid_list, msgtype='text', text={"content":content})
        _msg = json.dumps(_msg, ensure_ascii=False)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)
    
    @defer.inlineCallbacks       
    def create_qrcode(self,scene_id=None,expire_seconds=1800):
        _url = self.wx_create_qrcode()
        msg = dict(
            action_name="QR_SCENE",
            expire_seconds=expire_seconds,
            action_info = {'scene':{'scene_id':scene_id}}
        )
        _msg = json.dumps(msg, ensure_ascii=False)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

    @defer.inlineCallbacks
    def create_limit_qrcode(self,scene_str):
        _url = self.wx_create_qrcode()
        msg = dict(
            action_name="QR_LIMIT_STR_SCENE",
            action_info = {'scene':{'scene_str':scene_str}}
        )
        _msg = json.dumps(msg, ensure_ascii=False)
        _headers = {"Content-Type": "application/json;charset=utf-8"}
        _params = dict(method='POST', headers=_headers, use_gzip=True,
                       body=_msg, connect_timeout=300, request_timeout=300)
        resp = yield httpclient.fetch(_url, **_params)
        rdata = json.loads(resp.body)
        if resp.code != 200 or 'errcode' in rdata > 0:
            defer.returnValue(ApiMessage(code=resp.code, msg=resp.reason))
        else:
            apiresp = ApiMessage()
            apiresp.rdata=rdata
            apiresp.code=0
            apiresp.msg='success'
            defer.returnValue(apiresp)

