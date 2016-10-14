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
import traceback
import functools
from cyclone.util import ObjectDict
from toughlib import utils
from toughlib.paginator import Paginator
from toughradius import __version__ as sys_version
from toughlib.permit import permit
from toughradius.manage.settings import *
from toughradius import models
from toughlib import redis_session 
from toughlib import dispatch,logger
from twisted.python.failure import Failure

class BaseHandler(cyclone.web.RequestHandler):

    cache_key = 'toughradius.admin.'
    
    def __init__(self, *argc, **argkw):
        super(BaseHandler, self).__init__(*argc, **argkw)
        self.aes = self.application.aes
        self.cache = self.application.mcache
        self.db_backup = self.application.db_backup
        self.superrpc = self.application.superrpc
        self.session = redis_session.Session(self.application.session_manager, self)
        self.logtrace = self.application.logtrace

    def check_xsrf_cookie(self):
        if self.settings.config.system.get('production'):
            return super(BaseHandler, self).check_xsrf_cookie()

    def initialize(self):
        self.tp_lookup = self.application.tp_lookup
        self.db = self.application.db()
        
    def on_finish(self):
        self.db.close()
        
    def get_error_html(self, status_code=500, **kwargs):
        try:
            if 'exception' in kwargs:
                failure = kwargs.get("exception")
                if isinstance(failure, Failure):
                    logger.exception(failure.getTraceback())
                else:
                    logger.exception(failure)

            if self.request.headers.get('X-Requested-With') == 'XMLHttpRequest':
                return self.render_json(code=1, msg=u"%s:服务器处理失败，请联系管理员" % status_code)

            if status_code == 404:
                return self.render_string("error.html", msg=u"404:页面不存在")
            elif status_code == 403:
                return self.render_string("error.html", msg=u"403:非法的请求")
            elif status_code == 500:
                return self.render_string("error.html", msg=u"500:服务器处理失败，请联系管理员")
            else:
                return self.render_string("error.html", msg=u"%s:服务器处理失败，请联系管理员" % status_code)
        except Exception as err:
            logger.exception(err)
            return self.render_string("error.html", msg=u"%s:服务器处理失败，请联系管理员" % status_code)

    def render(self, template_name, **template_vars):
        html = self.render_string(template_name, **template_vars)
        self.write(html)

    def render_error(self, **template_vars):
        tpl = "error.html"
        html = self.render_string(tpl, **template_vars)
        self.write(html)

    def render_json(self, **template_vars):
        if not template_vars.has_key("code"):
            template_vars["code"] = 0
        resp = json.dumps(template_vars, ensure_ascii=False)
#        if self.settings.debug:
#            logger.debug("[api debug] :: %s response body: %s" % (self.request.path, utils.safeunicode(resp)))
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
        if self.current_user:
            template_vars["permit"] = self.current_user.permit
            template_vars["menu_icons"] = MENU_ICONS
            template_vars["all_menus"] = self.current_user.permit.build_menus(
                order_cats=ADMIN_MENUS
            )
        mytemplate = self.tp_lookup.get_template("admin/{0}".format(template_name))
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


    def set_session_user(self, username, ipaddr, opr_type, login_time):
        session_opr = ObjectDict()
        session_opr.username = username
        session_opr.ipaddr = ipaddr
        session_opr.opr_type = opr_type
        session_opr.login_time = login_time
        session_opr.resources = [r.rule_path for r in self.db.query(models.TrOperatorRule).filter_by(operator_name=username)]
        self.session['session_opr'] = session_opr
        self.session.save()

    def clear_session(self):
        self.session.clear()
        self.session.clear()
        self.clear_all_cookies()  
        
    def get_current_user(self):
        opr = self.session.get("session_opr")
        if opr:
            opr.permit = permit.fork(opr.username,opr.opr_type,opr.resources)
        return opr

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

    def get_opr_products(self):
        opr_type = int(self.current_user.opr_type)
        if opr_type == 0:
            return self.db.query(models.TrProduct).filter(
                models.TrProduct.product_status == 0,
                models.TrProduct.product_policy < FreeFee
            )
        else:
            return self.db.query(models.TrProduct).filter(
                models.TrProduct.id == models.TrOperatorProducts.product_id,
                models.TrOperatorProducts.operator_name == self.current_user.username,
                models.TrProduct.product_status == 0,
                models.TrProduct.product_policy < FreeFee
            )

    def get_opr_nodes(self):
        opr_type = int(self.current_user.opr_type)
        if opr_type == 0:
            return self.db.query(models.TrNode)
        opr_name = self.current_user.username
        return self.db.query(models.TrNode).filter(
            models.TrNode.node_name == models.TrOperatorNodes.node_name,
            models.TrOperatorNodes.operator_name == opr_name
        )

    def get_param_value(self, name, defval=None):
        val = self.db.query(models.TrParam.param_value).filter_by(param_name = name).scalar()
        return val or defval

    def add_oplog(self,message):
        ops_log = models.TrOperateLog()
        ops_log.operator_name = self.current_user.username
        ops_log.operate_ip = self.current_user.ipaddr
        ops_log.operate_time = utils.get_currtime()
        ops_log.operate_desc = message
        self.db.add(ops_log)

    def export_file(self, filename, data):
        self.set_header ('Content-Type', 'application/octet-stream')
        self.set_header ('Content-Disposition', 'attachment; filename=' + filename)
        self.write(data.xls)
        self.finish()

def authenticated(method):
    @functools.wraps(method)
    def wrapper(self, *args, **kwargs):
        if not self.current_user:
            if self.request.headers.get('X-Requested-With') == 'XMLHttpRequest': # jQuery 等库会附带这个头
                self.set_header('Content-Type', 'application/json; charset=UTF-8')
                self.write(json.dumps({'code': 1, 'msg': '您的会话已过期，请重新登录！'}))
                return
            if self.request.method in ("GET", "POST", "HEAD"):
                url = self.get_login_url()
                if "?" not in url:
                    if urlparse.urlsplit(url).scheme:
                        # if login url is absolute, make next absolute too
                        next_url = self.request.full_url()
                    else:
                        next_url = self.request.uri
                    url += "?" + urllib.urlencode(dict(next=next_url))
                self.redirect(url)
                return
            return self.render_error(msg=u"未授权的访问")
        else:
            if not self.current_user.permit.match(self.current_user.username,self.request.path):
                return self.render_error(msg=u"未授权的访问")
            return method(self, *args, **kwargs)
    return wrapper
