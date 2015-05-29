#!/usr/bin/env python
#coding:utf-8
import sys
import json
import re
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
import urlparse
import urllib
import time
import traceback
from cyclone.util import ObjectDict
from toughradius.console.libs import utils
from toughradius.console.libs.paginator import Paginator
from toughradius.console.libs.validate import vcache
from toughradius.console import models


class BaseHandler(cyclone.web.RequestHandler):
    
    def __init__(self, *argc, **argkw):
        super(BaseHandler, self).__init__(*argc, **argkw)
        self.logging = self.application.logging

    def initialize(self):
        self.db = self.application.db()
        self.tp_lookup = self.application.tp_lookup
        
    def on_finish(self):
        self.db.close()
        
    def get_error_html(self, status_code=500, **kwargs):
        if self.request.headers.get('X-Requested-With') == 'XMLHttpRequest':
            return self.render_json(code=1, msg=u"%s:服务器处理失败，请联系管理员" % status_code)

        if status_code == 404:
            return self.render_string("error.html", msg=u"404:页面不存在")
        elif status_code == 403:
            return self.render_string("error.html", msg=u"403:非法的请求")
        elif status_code == 500:
            if self.settings['debug']:
                return self.render_string("error.html", msg=traceback.format_exc())
            return self.render_string("error.html", msg=u"500:服务器处理失败，请联系管理员")
        else:
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
        self.write(resp)


    def render_string(self, template_name, **template_vars):
        template_vars["xsrf_form_html"] = self.xsrf_form_html
        template_vars["current_user"] = self.current_user
        template_vars["login_time"] = self.get_secure_cookie("portal_logintime")
        template_vars["request"] = self.request
        template_vars["handler"] = self
        template_vars["utils"] = utils
        mytemplate = self.tp_lookup.get_template(template_name)
        return mytemplate.render(**template_vars)


    def render_from_string(self, template_string, **template_vars):
        template = Template(template_string)
        return template.render(**template_vars)


    def get_page_data(self, query):
        page_size = self.application.settings.get("page_size",20)
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
        
    def get_current_user(self):
        username = self.get_secure_cookie("portal_user")
        if not username: return None
        ipaddr = self.request.remote_ip

        rad_session_user = self.db.query(models.SlcRadOnline).filter_by(
            account_number = username,
            framed_ipaddr = ipaddr
        ).first()

        if not rad_session_user:
            return None

        print rad_session_user

        raduser = self.db.query(models.SlcRadAccount).filter_by(
            account_number = username
        ).first()


        print raduser

        user = ObjectDict()
        user.username = username
        user.ipaddr = rad_session_user.framed_ipaddr
        user.times = rad_session_user.billing_times
        user.balance = raduser.balance

        print repr(user)
        return user


    def get_portal_name(self):
        param = self.db.query(models.SlcWlanParam).get("wlan_portal_name")
        return param.param_value

    def get_login_template(self,num=1):
        if self.chk_mobile():
            return "mobile_login_%s.html"%num
        else:
            return "login.html"

    def get_index_template(self,num=1):
        if self.chk_mobile():
            return "mobile_index_%s.html"%num
        else:
            return "index.html"


    def chk_mobile(self):
        try:
            userAgent = self.request.headers['User-Agent']
            # userAgent = env.get('HTTP_USER_AGENT')
            _long_matches = r'googlebot-mobile|android|avantgo|blackberry|blazer|elaine|hiptop|ip(hone|od)|kindle|midp|mmp|mobile|o2|opera mini|palm( os)?|pda|plucker|pocket|psp|smartphone|symbian|treo|up\.(browser|link)|vodafone|wap|windows ce; (iemobile|ppc)|xiino|maemo|fennec'
            _long_matches = re.compile(_long_matches, re.IGNORECASE)
            _short_matches = r'1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|e\-|e\/|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(di|rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|xda(\-|2|g)|yas\-|your|zeto|zte\-'
            _short_matches = re.compile(_short_matches, re.IGNORECASE)

            if _long_matches.search(userAgent) != None:
                return True
            user_agent = userAgent[0:4]
            if _short_matches.search(user_agent) != None:
                return True
            return False
        except:
            return False
    