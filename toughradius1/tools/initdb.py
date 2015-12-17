#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from toughradius.console import models
from toughradius.console.libs import utils
import argparse,ConfigParser
from sqlalchemy import *
from sqlalchemy.orm import scoped_session, sessionmaker
from hashlib import md5
from toughradius.console.libs import utils
import functools

def init_db(db):
    node = models.SlcNode()
    node.id = 1
    node.node_name = 'default'
    node.node_desc = u'测试区域'
    db.add(node)

    params = [
        ('system_name',u'管理系统名称',u'ToughRADIUS管理控制台'),
        ('customer_system_name',u'自助服务系统名称',u'ToughRADIUS自助服务中心'),
        ('customer_system_url',u"自助服务系统地址",u"http://forum.toughradius.net"),
        ('online_support', u"开启在线支持功能", u"1"),
        ('customer_must_active',u"激活邮箱才能自助开户充值(0:否|1:是)",u"0"),
        ('radiusd_address',u'Radius服务IP地址',u'127.0.0.1'),
        ('ticket_expire_days', u'上网日志保留天数', '90'),
        ('radiusd_admin_port',u'Radius服务管理端口',u'1815'),
        ('is_debug',u'DEBUG模式',u'0'),
        ('weixin_qrcode',u'微信公众号二维码图片(宽度230px)',u'http://img.toughradius.net/toughforum/jamiesun/1421820686.jpg!230'),
        ('service_phone',u'客户服务电话',u'000000'),
        ('service_qq',u'客户服务QQ号码',u'000000'),
        ('rcard_order_url',u'充值卡订购网站地址',u'http://www.tmall.com'),
        ('portal_secret',u'portal登陆密钥', u'abcdefg123456'),
        ('expire_notify_days','到期提醒提前天数',u'7'),
        ('expire_notify_tpl','到期提醒邮件模板',u'账号到期通知\n尊敬的会员您好:\n您的账号#account#即将在#expire#到期，请及时续费！'),
        ('expire_notify_url',u'到期通知url',u'http://your_notify_url?account={account}&expire={expire}&email={email}&mobile={mobile}'),
        ('expire_addrpool',u'到期提醒下发地址池',u'expire'),
        ('expire_session_timeout',u'到期用户下发最大会话时长(秒)',u'120'),
        ('smtp_server',u'SMTP服务器地址',u'smtp.mailgun.org'),
        ('smtp_user',u'SMTP用户名',u'service@toughradius.org'),
        ('smtp_pwd',u'SMTP密码',u'service2015'),
        ('smtp_sender',u'SMTP发送人名称',u'运营中心'),
        ('acct_interim_intelval',u'Radius记账间隔(秒)',u'120'),
        ('max_session_timeout',u'Radius最大会话时长(秒)',u'86400'),
        ('reject_delay',u'拒绝延迟时间(秒)(0-9)','0')
    ]

    for p in params:
        param = models.SlcParam()
        param.param_name = p[0]
        param.param_desc = p[1]
        param.param_value = p[2]
        db.add(param)

    opr = models.SlcOperator()
    opr.id = 1
    opr.operator_name = u'admin'
    opr.operator_type = 0
    opr.operator_pass = md5('root').hexdigest()
    opr.operator_desc = 'admin'
    opr.operator_status = 0
    db.add(opr)

    db.commit()
    db.close()

def update(db_engine):
    print 'starting update database...'
    metadata = models.get_metadata(db_engine)
    metadata.drop_all(db_engine)
    metadata.create_all(db_engine)
    print 'update database done'
    db = scoped_session(sessionmaker(bind=db_engine, autocommit=False, autoflush=True))()
    init_db(db)


        