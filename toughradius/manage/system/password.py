#!/usr/bin/env python
# coding:utf-8

from hashlib import md5

from toughlib import utils
from toughradius.manage.base import BaseHandler, MenuSys
from toughlib.permit import permit
from toughradius.manage import models
from toughradius.manage.system import password_forms
from toughradius.manage.settings import * 


###############################################################################
# password update
###############################################################################
@permit.route(r"/admin/password", u"密码修改", MenuSys, order=1.0100, is_menu=False)
class PasswordUpdateHandler(BaseHandler):
    def get(self):
        form = password_forms.password_update_form()
        form.fill(tr_user=self.current_user.username)
        return self.render("base_form.html", form=form)

    def post(self):
        form = password_forms.password_update_form()
        if not form.validates(source=self.get_params()):
            self.render("base_form.html", form=form)
            return
        if form.d.tr_user_pass != form.d.tr_user_pass_chk:
            self.render("base_form.html", form=form, msg=u'确认密码不一致')
            return
        opr = self.db.query(models.TrOperator).filter_by(operator_name=form.d.tr_user).first()
        opr.operator_pass = md5(form.d.tr_user_pass).hexdigest()

        self.add_oplog(u'修改%s密码 ' % (self.current_user.username))

        self.db.commit()
        self.redirect("/admin")


