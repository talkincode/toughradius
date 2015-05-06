#!/usr/bin/env python
#coding:utf-8
import sys
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
from toughradius.console.portal.base import BaseHandler
from toughradius.console.libs.apiutils import ApiMessage
from toughradius.console import models

class FetchPwdHandler(BaseHandler):

    def rand_number(self,is_pwd=False):
        r = ['0','1','2','3','4','5','6','7','8','9']
        rg = utils.random_generator
        def random_account():
            _num = ''.join([rg.choice(r) for _ in range(7)])
            if is_pwd:
                return _num
            if self.db.query(models.SlcRadAccount).filter_by(account_number=_num).count() > 0:
                return random_account()
            else:
                return _num

    def post(self):
        self.set_header("Content-Type", "application/json;charset=utf-8")
        body = self.request.body
        req = ApiMessage.parse(body)
        mp_openid = req.get_rval("open_id")
        mp_product_id＝ req.get_rval("product_id")
        mp_node_id＝ req.get_rval("node_id")
        member = self.db.query(models.SlcMember).filter_by(weixin_id=mp_openid)
        product = self.db.query(models.SlcRadProduct).get(mp_product_id)
        node = self.db.query(models.SlcNode).get(mp_node_id)

        if not product or not node:
            resp = ApiMessage(code=1,msg=u"产品套餐或区域节点不存在")
            self.write(resp.dumps())
            return


        account = None
        rand_ccount = self.rand_number()
        rand_pwd = self.rand_number(True)

        if member:
            account = self.db.query(models.SlcRadAccount).filter_by(account_number=member.member_name)
        else:
            member = models.SlcMember()
            member.node_id = mp_node_id
            member.realname = rand_ccount
            member.member_name = rand_ccount
            mpwd = rand_pwd
            member.password = md5(mpwd.encode()).hexdigest()
            member.idcard = mp_openid
            member.sex = '1'
            member.age = '0'
            member.email = ''
            member.mobile = mp_openid
            member.address = ''
            member.create_time = utils.get_currtime()
            member.update_time = utils.get_currtime()
            member.email_active = 0
            member.mobile_active = 0
            member.active_code = utils.get_uuid()
            self.db.add(member)
            self.db.flush()
            self.db.refresh(member)

            accept_log = models.SlcRadAcceptLog()
            accept_log.accept_type = 'open'
            accept_log.accept_source = 'weixin'
            accept_log.account_number = rand_ccount
            accept_log.accept_time = member.create_time
            accept_log.operator_name = get_cookie("username")
            accept_log.accept_desc = u"用户微信新开户：(%s)%s"%(member.member_name,member.realname)
            self.db.add(accept_log)
            self.db.flush()
            self.db.refresh(accept_log)

            order = models.SlcMemberOrder()
            order.order_id = utils.gen_order_id()
            order.member_id = member.member_id
            order.product_id = product.id
            order.account_number = rand_account
            order.order_fee = 0
            order.actual_fee = 0
            order.pay_status = 1
            order.accept_id = accept_log.id
            order.order_source = 'console'
            order.create_time = member.create_time
            order.order_desc = u"用户微信新开账号"
            self.db.add(order)

            account = models.SlcRadAccount()
            account.account_number = rand_account
            account.ip_address = ''
            account.member_id = member.member_id
            account.product_id = order.product_id
            account.install_address = member.address
            account.mac_addr = ''
            account.password = utils.encrypt(rand_pwd)
            account.status = 1
            account.balance = 0
            account.time_length = int(product.fee_times)
            account.flow_length = int(product.fee_flows)
            account.expire_date = expire_date
            account.user_concur_number = 1
            account.bind_mac = product.bind_mac
            account.bind_vlan = product.bind_vlan
            account.vlan_id = 0
            account.vlan_id2 = 0
            account.create_time = member.create_time
            account.update_time = member.create_time
            self.db.add(account)
            self.db.commit()

        resp = ApiMessage(code=0,msg=u"ok")
        resp.set_rval('account',account.account_number)
        resp.set_rval('password,'rand_pwd)
        self.write(resp.dumps())
        return




        
        