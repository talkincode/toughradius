#!/usr/bin/env python
#coding:utf-8
from hashlib import md5
from toughradius.common import utils
from toughradius.console.admin.base import BaseHandler
from toughradius.console import models
from toughradius.common.permit import permit

@permit.route(r"/login")
class LoginHandler(BaseHandler):

    def get(self):
        self.render("login.html")

    def post(self):
        uname = self.get_argument("username")
        upass = self.get_argument("password")
        if not uname:
            return self.render_json(code=1, msg=u"请填写用户名")
        if not upass:
            return self.render_json(code=1, msg=u"请填写密码")

        enpasswd = md5(upass.encode()).hexdigest()

        opr = self.db.query(models.TrOperator).filter_by(
            operator_name=uname,
            operator_pass=enpasswd
        ).first()
        if not opr:
            return self.render_json(code=1, msg=u"用户名密码不符")

        if opr.operator_status == 1:
            return self.render_json(code=1, msg=u"该操作员账号已被停用")

        self.set_secure_cookie("tr_user", uname, expires_days=None)
        self.set_secure_cookie("tr_login_time", utils.get_currtime(), expires_days=None)
        self.set_secure_cookie("tr_login_ip", self.request.remote_ip, expires_days=None)
        self.set_secure_cookie("tr_opr_type", str(opr.operator_type), expires_days=None)

        if opr.operator_type == 1:
            for rule in self.db.query(models.TrOperatorRule).filter_by(operator_name=uname):
                permit.bind_opr(rule.operator_name, rule.rule_path)

        ops_log = models.TrOperateLog()
        ops_log.operator_name = uname
        ops_log.operate_ip = self.request.remote_ip
        ops_log.operate_time = utils.get_currtime()
        ops_log.operate_desc = u'操作员(%s)登陆' % (uname,)
        self.db.add(ops_log)
        self.db.commit()

        self.render_json(code=0, msg="ok")

