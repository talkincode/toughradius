#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import datetime
from toughradius import models
from toughradius.manage.base import BaseHandler
from toughradius.manage.resource import product_forms
from toughradius.common.permit import permit
from toughradius.common import utils
from toughradius import settings 


@permit.route(r"/admin/product", u"资费套餐管理",settings.MenuRes, order=3.0000, is_menu=True)
class ProductListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        query = self.get_opr_products()
        self.render(
            "product_list.html",
            product_policys=product_forms.product_policy,
            node_list=self.db.query(models.TrNode),
            page_data=self.get_page_data(query)
        )

@permit.route(r"/admin/product/add", u"新增资费套餐",settings.MenuRes, order=3.0001)
class ProductAddListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.render("product_form.html", form=product_forms.product_add_form())

    @cyclone.web.authenticated
    def post(self):
        form = product_forms.product_add_form()
        if not form.validates(source=self.get_params()):
            return self.render("product_form.html", form=form)

        product = models.TrProduct()
        product.product_name = form.d.product_name
        product.product_policy = form.d.product_policy
        product.product_status = form.d.product_status
        product.fee_months = int(form.d.get("fee_months", 0))
        product.fee_times = utils.hour2sec(form.d.get("fee_times", 0))
        product.fee_flows = utils.mb2kb(form.d.get("fee_flows", 0))
        product.bind_mac = form.d.bind_mac
        product.bind_vlan = form.d.bind_vlan
        product.concur_number = form.d.concur_number
        product.fee_price = utils.yuan2fen(form.d.fee_price)
        product.fee_period =  '' #form.d.fee_period or ''
        product.input_max_limit = utils.mbps2bps(form.d.input_max_limit)
        product.output_max_limit = utils.mbps2bps(form.d.output_max_limit)
        _datetime = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        product.create_time = _datetime
        product.update_time = _datetime
        self.db.add(product)
        self.add_oplog(u'新增资费信息:%s' % utils.safeunicode(product.product_name))
        self.db.commit()
        self.redirect("/admin/product", permanent=False)

@permit.route(r"/admin/product/update", u"修改资费套餐",settings.MenuRes, order=3.0002)
class ProductUpdateListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        product_id = self.get_argument("product_id")
        form = product_forms.product_update_form()
        product = self.db.query(models.TrProduct).get(product_id)
        form.fill(product)
        form.product_policy_name.set_value(product_forms.product_policy[product.product_policy])
        form.fee_times.set_value(utils.sec2hour(product.fee_times))
        form.fee_flows.set_value(utils.kb2mb(product.fee_flows))
        form.input_max_limit.set_value(utils.bps2mbps(product.input_max_limit))
        form.output_max_limit.set_value(utils.bps2mbps(product.output_max_limit))
        form.fee_price.set_value(utils.fen2yuan(product.fee_price))
        return self.render("product_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        form = product_forms.product_update_form()
        if not form.validates(source=self.get_params()):
            return self.render("product_form.html", form=form)

        product = self.db.query(models.TrProduct).get(form.d.id)
        product.product_name = form.d.product_name
        product.product_status = form.d.product_status
        product.fee_months = int(form.d.get("fee_months", 0))
        product.fee_times = utils.hour2sec(form.d.get("fee_times", 0))
        product.fee_flows = utils.mb2kb(form.d.get("fee_flows", 0))
        product.bind_mac = form.d.bind_mac
        product.bind_vlan = form.d.bind_vlan
        product.concur_number = form.d.concur_number
        product.fee_period = ''#form.d.fee_period or ''
        product.fee_price = utils.yuan2fen(form.d.fee_price)
        product.input_max_limit = utils.mbps2bps(form.d.input_max_limit)
        product.output_max_limit = utils.mbps2bps(form.d.output_max_limit)
        product.update_time = utils.get_currtime()
        self.add_oplog(u'修改资费信息:%s' % utils.safeunicode(product.product_name))
        self.db.commit()
        self.redirect("/admin/product", permanent=False)


@permit.route(r"/admin/product/delete", u"删除资费套餐",settings.MenuRes, order=3.0003)
class ProductDeleteListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        product_id = self.get_argument("product_id")
        if self.db.query(models.TrAccount).filter_by(product_id=product_id).count() > 0:
            return self.render_error(msg=u"该套餐有用户使用，不允许删除")

        self.db.query(models.TrProduct).filter_by(id=product_id).delete()
        self.add_oplog(u'删除资费信息:%s' % product_id)
        self.db.commit()
        self.redirect("/admin/product", permanent=False)

@permit.route(r"/admin/product/detail", u"资费详情",settings.MenuRes, order=3.0004)
class ProductDeleteListHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        product_id = self.get_argument("product_id")
        product = self.db.query(models.TrProduct).get(product_id)
        if not product:
            return self.render_error(msg=u"资费不存在")

        product_attrs = self.db.query(models.TrProductAttr).filter_by(product_id=product_id)
        return self.render("product_detail.html",
                      product_policys=product_forms.product_policy,
                      product=product, product_attrs=product_attrs)









