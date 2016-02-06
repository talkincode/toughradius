#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web

from toughradius.manage import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.resource import bas_forms
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 

@permit.route(r"/admin/bas", u"接入设备管理",MenuRes, order=2.0000, is_menu=True)
class BasListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.render("bas_list.html",
                  bastype=bas_forms.bastype,
                  bas_list=self.db.query(models.TrBas))

@permit.route(r"/admin/bas/add", u"新增接入设备", MenuRes, order=2.0001)
class BasAddHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        form = bas_forms.bas_add_form()
        self.render("base_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        form = bas_forms.bas_add_form()
        if not form.validates(source=self.get_params()):
            return self.render("base_form.html", form=form)

        if not any([form.d.ip_addr,form.d.dns_name]):
            return self.render("base_form.html", form=form, msg=u"ip地址或域名至少填写一项")

        if self.db.query(models.TrBas.id).filter_by(ip_addr=form.d.ip_addr).count() > 0:
            return self.render("base_form.html", form=form, msg=u"接入设备地址已经存在")

        bas = models.TrBas()
        bas.ip_addr = form.d.ip_addr
        bas.dns_name = form.d.dns_name
        bas.bas_name = form.d.bas_name
        bas.time_type = form.d.time_type
        bas.vendor_id = form.d.vendor_id
        bas.bas_secret = form.d.bas_secret
        bas.coa_port = form.d.coa_port
        self.db.add(bas)

        self.add_oplog(u'新增接入设备信息:%s' % bas.ip_addr)

        self.db.commit()
        self.redirect("/admin/bas",permanent=False)

@permit.route(r"/admin/bas/update", u"修改接入设备", MenuRes, order=2.0002)
class BasUpdateHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        bas_id = self.get_argument("bas_id")
        form = bas_forms.bas_update_form()
        form.fill(self.db.query(models.TrBas).get(bas_id))
        self.render("base_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        form = bas_forms.bas_update_form()
        if not form.validates(source=self.get_params()):
            return self.render("base_form.html", form=form)
        bas = self.db.query(models.TrBas).get(form.d.id)
        bas.dns_name = form.d.dns_name
        bas.bas_name = form.d.bas_name
        bas.time_type = form.d.time_type
        bas.vendor_id = form.d.vendor_id
        bas.bas_secret = form.d.bas_secret
        bas.coa_port = form.d.coa_port

        self.add_oplog(u'修改接入设备信息:%s' % bas.ip_addr)

        self.db.commit()
        self.redirect("/admin/bas",permanent=False)


@permit.route(r"/admin/bas/delete", u"删除接入设备", MenuRes, order=2.0003)
class BasDeleteHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        bas_id = self.get_argument("bas_id")
        self.db.query(models.TrBas).filter_by(id=bas_id).delete()

        self.add_oplog(u'删除接入设备信息:%s' % bas_id)

        self.db.commit()
        self.redirect("/admin/bas",permanent=False)
