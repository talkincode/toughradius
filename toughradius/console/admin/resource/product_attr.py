#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web

from toughradius.console import models
from toughradius.console.admin.base import BaseHandler
from toughradius.console.admin.resource import product_forms
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius.common.radius_attrs import radius_attrs 
from toughradius.common.settings import * 



@permit.route(r"/product/attr/add", u"新增资费属性",MenuRes, order=3.0001)
class ProductAddListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        product_id = self.get_argument("product_id")
        if self.db.query(models.TrProduct).filter_by(id=product_id).count() <= 0:
            return self.render_error(msg=u"资费不存在")
        form = product_forms.product_attr_add_form()
        form.product_id.set_value(product_id)
        return self.render("pattr_form.html", form=form, pattrs=radius_attrs)

    def post(self):
        form = product_forms.product_attr_add_form()
        if not form.validates(source=self.get_params()):
            return self.render("pattr_form.html", form=form, pattrs=radius_attrs)

        attr = models.TrProductAttr()
        attr.product_id = form.d.product_id
        attr.attr_type = 1
        attr.attr_name = form.d.attr_name
        attr.attr_value = form.d.attr_value
        attr.attr_desc = form.d.attr_desc
        self.db.add(attr)
        self.add_oplog(u'新增资费属性信息:%s' %  attr.attr_name)
        self.db.commit()

        self.redirect("/product/detail?product_id=%s" % form.d.product_id)

@permit.route(r"/product/attr/update", u"修改资费属性",MenuRes, order=3.0002)
class ProductUpdateListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        attr_id = self.get_argument("attr_id")
        attr = self.db.query(models.TrProductAttr).get(attr_id)
        form = product_forms.product_attr_update_form()
        form.fill(attr)
        return self.render("pattr_form.html", form=form, pattrs=radius_attrs)

    def post(self):
        form = product_forms.product_attr_update_form()
        if not form.validates(source=self.get_params()):
            return self.render("pattr_form.html", form=form, pattrs=radius_attrs)

        attr = self.db.query(models.TrProductAttr).get(form.d.id)
        attr.attr_type = 1
        attr.attr_name = form.d.attr_name
        attr.attr_value = form.d.attr_value
        attr.attr_desc = form.d.attr_desc
        self.add_oplog(u'修改资费属性信息:%s' % attr.attr_name)
        self.db.commit()
        self.redirect("/product/detail?product_id=%s" % form.d.product_id)

@permit.route(r"/product/attr/delete", u"删除资费属性",MenuRes, order=3.0003)
class ProductDeleteListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        attr_id = self.get_argument("attr_id")
        attr = self.db.query(models.TrProductAttr).get(attr_id)
        product_id = attr.product_id
        self.db.query(models.TrProductAttr).filter_by(id=attr_id).delete()
        self.add_oplog(u'删除资费属性信息:%s' % attr.attr_name)
        self.db.commit()
        self.redirect("/product/detail?product_id=%s" % product_id)









