#!/usr/bin/env python
# coding=utf-8
import cyclone.auth
import cyclone.escape
import cyclone.web
from toughradius.manage.base import BaseHandler
from toughradius.manage.system import param_forms
from toughradius import models
from toughradius.common.permit import permit
from toughradius.common import dispatch,redis_cache
from toughradius import settings 

@permit.route("/admin/param", u"系统参数管理", settings.MenuSys, is_menu=True, order=2.0005)
class ParamHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        active = self.get_argument("active","syscfg")
        sys_form = param_forms.sys_form()
        notify_form = param_forms.notify_form()
        mail_form = param_forms.mail_form()
        rad_form = param_forms.rad_form()
        fparam = {}
        for p in self.db.query(models.TrParam):
            fparam[p.param_name] = p.param_value

        for form in (sys_form, notify_form, mail_form, rad_form):
            form.fill(fparam)

        return self.render("param.html",
                      active=active,
                      sys_form=sys_form,
                      mail_form=mail_form,
                      notify_form=notify_form,
                      rad_form=rad_form)


@permit.route("/admin/param/update", u"系统参数更新", settings.MenuSys, order=2.0006)
class ParamUpdateHandler(BaseHandler):

    @cyclone.web.authenticated
    def post(self):
        active = self.get_argument("active", "syscfg")
        for param_name in self.get_params():
            if param_name in ("active", "submit"):
                continue

            param = self.db.query(models.TrParam).filter_by(param_name=param_name).first()
            if not param:
                param = models.TrParam()
                param.param_name = param_name
                param.param_value = self.get_argument(param_name)
                self.db.add(param)
            else:
                param.param_value = self.get_argument(param_name)

            dispatch.pub(redis_cache.CACHE_SET_EVENT,param.param_name,param.param_value,600)

        self.add_oplog(u'操作员(%s)修改参数' % (self.current_user.username))
        self.db.commit()
        self.redirect("/admin/param?active=%s" % active)
