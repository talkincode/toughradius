#!/usr/bin/env python
#coding:utf-8
import json
import re
import urlparse
import urllib
import traceback
import cyclone.auth
import cyclone.escape
import cyclone.web
import tempfile
import functools
from cyclone.util import ObjectDict
from toughlib import utils
from toughlib.paginator import Paginator
from toughradius import __version__ as sys_version
from toughlib.permit import permit
from toughradius.manage.settings import *
from toughradius.manage import models
from toughlib import db_session as session

def authenticated(method):
    """ 登陆校验装饰器 """

    @functools.wraps(method)
    def wrapper(self, *args, **kwargs):
        if not self.current_user:
            if self.request.headers.get('X-Requested-With') == 'XMLHttpRequest': # jQuery 等库会附带这个头
                self.set_header('Content-Type', 'application/json; charset=UTF-8')
                self.write(json.dumps({'code': 1, 'msg': u'您的会话已过期，请重新登录！'}))
                return
            if self.request.method in ("GET", "POST", "HEAD"):
                url = "/customer/login"
                if "?" not in url:
                    if urlparse.urlsplit(url).scheme:
                        # if login url is absolute, make next absolute too
                        next_url = self.request.full_url()
                    else:
                        next_url = self.request.uri
                    url += "?" + urllib.urlencode(dict(next=next_url))
                self.redirect(url)
                return
            raise cyclone.web.HTTPError(403)
        else:
            # execute method and process error
            try:
                return method(self, *args, **kwargs)
            except Exception,e:
                traceback.print_exc()
                if self.request.headers.get('X-Requested-With') == 'XMLHttpRequest': # jQuery 等库会附带这个头
                    self.set_header('Content-Type', 'application/json; charset=UTF-8')
                    self.write(json.dumps({'code': 1, 'msg': u'服务器处理失败,请稍后再试！'}))
                else:
                    self.render_error(msg=u"服务器处理失败,请稍后再试")
    return wrapper

class BaseHandler(cyclone.web.RequestHandler):
    
    def __init__(self, *argc, **argkw):
        super(BaseHandler, self).__init__(*argc, **argkw)
        self.syslog = self.application.syslog
        self.aes = self.application.aes
        self.cache = self.application.mcache
        self.session = session.Session(self.application.session_manager, self)

    def initialize(self):
        self.tp_lookup = self.application.tp_lookup
        self.db = self.application.db()
        
    def on_finish(self):
        self.db.close()
        
    def get_error_html(self, status_code=500, **kwargs):
        # self.syslog.error("http error : [status_code:{0}], {1}".format(status_code, utils.safestr(kwargs)))
        if self.settings.debug:
            traceback.print_exc()
        if self.request.headers.get('X-Requested-With') == 'XMLHttpRequest':
            return self.render_json(code=1, msg=u"%s:服务器处理失败，请联系管理员" % status_code)

        if status_code == 404:
            return self.render_error(msg=u"404:页面不存在")
        elif status_code == 403:
            return self.render_error(msg=u"403:非法的请求")
        elif status_code == 500:
            return self.render_error(msg=u"500:服务器处理失败，请联系管理员")
        else:
            return self.render_error(msg=u"%s:服务器处理失败，请联系管理员" % status_code)

    def render(self, template_name, **template_vars):
        html = self.render_string("ssportal/%s" % template_name, **template_vars)
        self.write(html)

    def render_error(self, **template_vars):
        tpl = "ssportal/error.html"
        html = self.render_string(tpl, **template_vars)
        self.write(html)

    def render_json(self, **template_vars):
        if not template_vars.has_key("code"):
            template_vars["code"] = 0
        resp = json.dumps(template_vars, ensure_ascii=False)
        if self.settings.debug:
            self.syslog.debug("[api debug] :: %s response body: %s" % (self.request.path, utils.safeunicode(resp)))
        self.write(resp)


    def render_string(self, template_name, **template_vars):
        template_vars["xsrf_form_html"] = self.xsrf_form_html
        template_vars["current_user"] = self.current_user
        template_vars["login_time"] = self.get_secure_cookie("tr_login_time")
        template_vars["request"] = self.request
        template_vars["requri"] = "{0}://{1}".format(self.request.protocol, self.request.host)
        template_vars["handler"] = self
        template_vars["utils"] = utils
        template_vars['sys_version'] = sys_version
        mytemplate = self.tp_lookup.get_template(template_name)
        return mytemplate.render(**template_vars)


    def render_from_string(self, template_string, **template_vars):
        from mako.template import Template
        template = Template(template_string)
        return template.render(**template_vars)


    def get_page_data(self, query):
        page_size = self.application.settings.get("page_size",15)
        page = int(self.get_argument("page", 1))
        offset = (page - 1) * page_size
        result = query.limit(page_size).offset(offset)
        page_data = Paginator(self.get_page_url, page, query.count(), page_size)
        page_data.result = result
        return page_data
   

    def get_page_url(self, page, form_id=None):
        if form_id:
            return "javascript:goto_page('%s',%s);" %(form_id.strip(),page)
        path = self.request.path
        query = self.request.query
        qdict = urlparse.parse_qs(query)
        for k, v in qdict.items():
            if isinstance(v, list):
                qdict[k] = v and v[0] or ''

        qdict['page'] = page
        return path + '?' + urllib.urlencode(qdict)

    def set_session_user(self, customer, ipaddr, login_time):
        session_user = ObjectDict()
        session_user.username = customer.customer_name
        session_user.cid = customer.customer_id
        session_user.ipaddr = ipaddr
        session_user.login_time = login_time
        self.session['session_user'] = session_user
        self.session.save()

    def clear_session(self):
        self.session.clear()
        self.session.save()
        self.clear_all_cookies()  
        
    def get_current_user(self):
        return self.session.get("session_user")

    def get_params(self):
        arguments = self.request.arguments
        params = {}
        for k, v in arguments.items():
            if len(v) == 1:
                params[k] = v[0]
            else:
                params[k] = v
        return params

    def get_params_obj(self, obj):
        arguments = self.request.arguments
        for k, v in arguments.items():
            if len(v) == 1:
                if type(v[0]) == str:
                    setattr(obj, k, v[0].decode('utf-8', ''))
                else:
                    setattr(obj, k, v[0])
            else:
                if type(v) == str:
                    setattr(obj, k, v.decode('utf-8'))
                else:
                    setattr(obj, k, v)
        return obj

    # service method

    def get_online_status(self, account_number):
        return self.db.query(models.TrOnline.id).filter_by(account_number = account_number).count() > 0


    def get_param_value(self, name, defval=None):
        val = self.db.query(models.TrParam.param_value).filter_by(param_name = name).scalar()
        return val or defval


    def export_file(self, filename, data):
        self.set_header ('Content-Type', 'application/octet-stream')
        self.set_header ('Content-Disposition', 'attachment; filename=' + filename)
        self.write(data.xls)
        self.finish()






