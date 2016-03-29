#!/usr/bin/env python
#coding=utf-8

import cyclone.auth
import cyclone.escape
import cyclone.web
import decimal
import datetime
from tablib import Dataset
from hashlib import md5
from toughradius.manage import models
from toughradius.manage.customer import customer_forms
from toughradius.manage.customer.customer import CustomerHandler
from toughlib.permit import permit
from toughlib import utils
from toughradius.manage.settings import * 


@permit.route(r"/admin/customer/import", u"用户资料导入",MenuUser, order=1.3000, is_menu=True)
class CustomerImportHandler(CustomerHandler):

    @cyclone.web.authenticated
    def get(self):
        nodes = [(n.id, n.node_desc) for n in self.get_opr_nodes()]
        products = [(p.id, p.product_name) for p in self.get_opr_products()]
        form = customer_forms.customer_import_form(nodes, products)
        return self.render("customer_import_form.html", form=form)

    @cyclone.web.authenticated
    def post(self):
        nodes = [(n.id, n.node_desc) for n in self.get_opr_nodes()]
        products = [(n.id, n.product_name) for n in self.get_opr_products()]
        iform = customer_forms.customer_import_form(nodes, products)
        node_id = self.get_argument('node_id')
        product_id = self.get_argument('product_id')
        f = self.request.files['import_file'][0]
        try:
            impctx = utils.safeunicode(f['body'])
        except Exception as err:
            self.render_error(msg=u"File format error： %s" % utils.safeunicode(err))
            return
        lines = impctx.split("\n")
        _num = 0
        impusers = []
        for line in lines:
            _num += 1
            line = line.strip()
            if not line or u"用户姓名" in line: continue
            attr_array = line.split(",")
            if len(attr_array) < 11:
                return self.render("customer_import_form.html", form=iform, msg=u"line %s error: length must 11 " % _num)

            vform = customer_forms.customer_import_vform()
            if not vform.validates(dict(
                realname=attr_array[0],
                idcard=attr_array[1],
                mobile=attr_array[2],
                address=attr_array[3],
                account_number=attr_array[4],
                password=attr_array[5],
                begin_date=attr_array[6],
                expire_date=attr_array[7],
                balance=attr_array[8],
                time_length=utils.hour2sec(attr_array[9]),
                flow_length=utils.mb2kb(attr_array[10]))):
                return self.render("customer_import_form.html", form=iform, msg=u"line %s error: %s" % (_num, vform.errors))

            impusers.append(vform)

        _unums = 0
        for form in impusers:
            try:
                customer = models.TrCustomer()
                customer.node_id = node_id
                customer.realname = form.d.realname
                customer.idcard = form.d.idcard
                customer.customer_name = form.d.account_number
                customer.password = md5(form.d.password.encode()).hexdigest()
                customer.sex = '1'
                customer.age = '0'
                customer.email = ''
                customer.mobile = form.d.mobile
                customer.address = form.d.address
                customer.create_time = form.d.begin_date + ' 00:00:00'
                customer.update_time = utils.get_currtime()
                customer.email_active = 0
                customer.mobile_active = 0
                customer.active_code = utils.get_uuid()
                self.db.add(customer)
                self.db.flush()
                self.db.refresh(customer)

                accept_log = models.TrAcceptLog()
                accept_log.accept_type = 'open'
                accept_log.accept_source = 'console'
                _desc = u"用户导入账号：%s" % form.d.account_number
                accept_log.accept_desc = _desc
                accept_log.account_number = form.d.account_number
                accept_log.accept_time = customer.update_time
                accept_log.operator_name = self.current_user.username
                self.db.add(accept_log)
                self.db.flush()
                self.db.refresh(accept_log)

                order_fee = 0
                actual_fee = 0
                balance = 0
                time_length = 0
                flow_length = 0
                expire_date = form.d.expire_date
                product = self.db.query(models.TrProduct).get(product_id)
                # 买断时长
                if product.product_policy == BOTimes:
                    time_length = int(form.d.time_length)
                # 买断流量
                elif product.product_policy == BOFlows:
                    flow_length = int(form.d.flow_length)
                # 预付费时长,预付费流量
                elif product.product_policy in (PPTimes, PPFlow):
                    balance = utils.yuan2fen(form.d.balance)
                    expire_date = MAX_EXPIRE_DATE

                order = models.TrCustomerOrder()
                order.order_id = utils.gen_order_id()
                order.customer_id = customer.customer_id
                order.product_id = product.id
                order.account_number = form.d.account_number
                order.order_fee = order_fee
                order.actual_fee = actual_fee
                order.pay_status = 1
                order.accept_id = accept_log.id
                order.order_source = 'console'
                order.create_time = customer.update_time
                order.order_desc = u"用户导入开户"
                self.db.add(order)

                account = models.TrAccount()
                account.account_number = form.d.account_number
                account.customer_id = customer.customer_id
                account.product_id = order.product_id
                account.install_address = customer.address
                account.ip_address = ''
                account.mac_addr = ''
                account.password = self.aes.encrypt(form.d.password)
                account.status = 1
                account.balance = balance
                account.time_length = time_length
                account.flow_length = flow_length
                account.expire_date = expire_date
                account.user_concur_number = product.concur_number
                account.bind_mac = product.bind_mac
                account.bind_vlan = product.bind_vlan
                account.vlan_id1 = 0
                account.vlan_id2 = 0
                account.create_time = customer.create_time
                account.update_time = customer.update_time
                self.db.add(account)
                _unums += 1

            except Exception as e:
                return self.render("customer_import_form.html", form=iform, msg=u"error : %s" % str(e))

        self.add_oplog(u"导入开户，用户数：%s" % _unums)
        self.db.commit()
        self.redirect("/admin/customer")