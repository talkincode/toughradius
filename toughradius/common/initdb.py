#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys
import os
import time
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from toughlib import utils
from toughradius.manage import models
from toughlib.dbengine import get_engine
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.manage.settings import FREE_FEE_PID, FreeFee
from hashlib import md5


def init_db(db):
    node = models.TrNode()
    node.id = 1
    node.node_name = 'default'
    node.node_desc = u'默认区域'
    db.add(node)

    params = [
        ('system_name', u'管理系统名称', u'ToughRADIUS管理控制台'),
        ('system_ticket_expire_days', u'上网日志保留天数', '30'),
        ('is_debug', u'DEBUG模式', u'0'),
        ('expire_notify_days', '到期提醒提前天数', u'7'),
        ('expire_notify_interval', '到期提醒提前间隔(分钟)', u'1440'),
        ('smtp_notify_tpl', '到期提醒邮件模板', u'账号到期通知\n尊敬的会员您好:\n您的账号#account#即将在#expire#到期，请及时续费！'),
        ('expire_notify_url', u'到期通知url', u'http://127.0.0.1:1816?account={account}&expire={expire}&email={email}&mobile={mobile}'),
        ('expire_notify_time', u'到期通知时间', '09:00'),
        ('expire_addrpool', u'到期提醒下发地址池', u'expire'),
        ('expire_session_timeout', u'到期用户下发最大会话时长(秒)', u'120'),
        ('mail_notify_enable', u'启动邮件到期提醒', u'0'),
        ('sms_notify_enable', u'启动短信到期提醒', u'0'),
        ('webhook_notify_enable', u'启动URL触发到期提醒', u'0'),
        ('mail_mode', u'邮件通知服务类型', u'toughcloud'),
        ('smtp_server', u'SMTP服务器地址', u'smtp.mailgun.org'),
        ('smtp_port', u'SMTP服务器端口', u'25'),
        ('smtp_user', u'SMTP用户名', u'service@toughradius.org'),
        ('smtp_pwd', u'SMTP密码', u'service2015'),
        ('smtp_sender', u'SMTP发送人名称', u'运营中心'),
        ('smtp_from', u'SMTP邮件发送地址', u'service@toughradius.org'),
        ('radius_bypass', u'Radius认证密码模式', u'1'),
        ('radius_acct_interim_intelval', u'Radius记账间隔(秒)', u'300'),
        ('radius_max_session_timeout', u'Radius最大会话时长(秒)', u'86400'),
        ('radius_auth_auto_unlock', u'并发自动解锁', '0'),
        ('radius_user_trace', u'跟踪用户 Radius 消息', '0'),
    ]

    for p in params:
        param = models.TrParam()
        param.param_name = p[0]
        param.param_desc = p[1]
        param.param_value = p[2]
        db.add(param)

    opr = models.TrOperator()
    opr.id = 1
    opr.operator_name = u'admin'
    opr.operator_type = 0
    opr.operator_pass = md5('root').hexdigest()
    opr.operator_desc = 'admin'
    opr.operator_status = 0
    db.add(opr)

    bas = models.TrBas()
    bas.ip_addr = '127.0.0.1'
    bas.vendor_id = '0'
    bas.bas_name = 'local bras'
    bas.bas_secret = 'secret'
    bas.coa_port = 3799
    bas.time_type = 0
    db.add(bas)

    free_product = models.TrProduct()
    free_product.id = FREE_FEE_PID
    free_product.product_name = u"自由资费"
    free_product.product_policy = FreeFee
    free_product.product_status = 0
    free_product.fee_months = 0
    free_product.fee_times = 0
    free_product.fee_flows = 0
    free_product.bind_mac = 0
    free_product.bind_vlan = 0
    free_product.concur_number = 0
    free_product.fee_price = 0
    free_product.fee_period =  '' 
    free_product.input_max_limit = 0
    free_product.output_max_limit = 0
    free_product.create_time = utils.get_currtime()
    free_product.update_time = utils.get_currtime()
    db.add(free_product)

    db.commit()
    db.close()

def update(config,force=False):
    try:
        db_engine = get_engine(config)
        if int(os.environ.get("DB_INIT", 1)) == 1 or force:
            print 'starting update database...'
            metadata = models.get_metadata(db_engine)
            metadata.drop_all(db_engine)
            metadata.create_all(db_engine)
            print 'update database done'
            db = scoped_session(sessionmaker(bind=db_engine, autocommit=False, autoflush=True))()
            init_db(db)
    except:
        import traceback
        traceback.print_exc()



