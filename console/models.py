#!/usr/bin/env python
#coding:utf-8

from sqlalchemy import *
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relation
from sqlalchemy.orm import scoped_session, sessionmaker
from hashlib import md5
from libs import utils

engine = create_engine('mysql://root:root@127.0.0.1:3306/mysql?charset=utf8')
DeclarativeBase = declarative_base()
metadata = DeclarativeBase.metadata
metadata.bind = engine
 
class SlcNode(DeclarativeBase):
    """区域表"""
    __tablename__ = 'slc_node'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"区域编号")
    node_name = Column(u'node_name', VARCHAR(length=32), nullable=False,doc=u"区域名")
    node_desc = Column(u'node_desc', VARCHAR(length=64), nullable=False,doc=u"区域描述")

    #relation definitions

class SlcOperator(DeclarativeBase):
    """操作员表 操作员类型 0 系统管理员 1 普通操作员"""
    __tablename__ = 'slc_rad_operator'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"操作员id")
    node_id = Column('node_id', INTEGER(), nullable=False,doc=u"操作员区域")
    operator_type = Column('operator_type', INTEGER(), nullable=False,doc=u"操作员类型")
    operator_name = Column(u'operator_name', VARCHAR(32), nullable=False,doc=u"操作员名称")
    operator_pass = Column(u'operator_pass', VARCHAR(length=128), nullable=False,doc=u"操作员密码")
    operator_status = Column(u'operator_status', INTEGER(), nullable=False,doc=u"操作员状态,0/1")
    operator_desc = Column(u'operator_desc', VARCHAR(255), nullable=False,doc=u"操作员描述")    

class SlcParam(DeclarativeBase):
    """系统参数表  <radiusd default table>"""
    __tablename__ = 'slc_param'

    __table_args__ = {}

    #column definitions
    param_name = Column(u'param_name', VARCHAR(length=64), primary_key=True, nullable=False,doc=u"参数名")
    param_value = Column(u'param_value', VARCHAR(length=255), nullable=False,doc=u"参数值")
    param_desc = Column(u'param_desc', VARCHAR(length=255),doc=u"参数描述")

    #relation definitions

class SlcRadBas(DeclarativeBase):
    """BAS设备表 <radiusd default table>"""
    __tablename__ = 'slc_rad_bas'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"设备id")
    vendor_id = Column(u'vendor_id', VARCHAR(length=32), nullable=False,doc=u"厂商标识")
    ip_addr = Column(u'ip_addr', VARCHAR(length=15), nullable=False,doc=u"IP地址")
    bas_name = Column(u'bas_name', VARCHAR(length=64), nullable=False,doc=u"bas名称")
    bas_secret = Column(u'bas_secret', VARCHAR(length=64), nullable=False,doc=u"共享密钥")
    coa_port = Column(u'coa_port', INTEGER(), nullable=False,doc=u"CoA端口")
    time_type = Column(u'time_type', SMALLINT(), nullable=False,doc=u"时区类型")

    #relation definitions


class SlcRadGroup(DeclarativeBase):
    """认证策略用户组 <radiusd default table>"""
    __tablename__ = 'slc_rad_group'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"用户组id")
    group_name = Column(u'group_name', VARCHAR(length=64), nullable=False,doc=u"用户组名")
    group_desc = Column(u'group_desc', VARCHAR(length=255),doc=u"用户组描述")
    bind_mac = Column(u'bind_mac', SMALLINT(), nullable=False,doc=u"是否绑定mac")
    bind_vlan = Column(u'bind_vlan', SMALLINT(), nullable=False,doc=u"是否绑定vlan")
    concur_number = Column(u'concur_number', INTEGER(), nullable=False,doc=u"并发数")
    update_time = Column(u'update_time', VARCHAR(length=19), nullable=False,doc=u"更新时间")

    #relation definitions


class SlcRadRoster(DeclarativeBase):
    """黑白名单 0 白名单 1 黑名单 <radiusd default table>"""
    __tablename__ = 'slc_rad_roster'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"黑白名单id")
    mac_addr = Column('mac_addr', VARCHAR(length=17), nullable=False,doc=u"mac地址")
    account_number = Column('account_number', VARCHAR(length=32),doc=u"上网账号")
    begin_time = Column('begin_time', VARCHAR(length=19), nullable=False,doc=u"生效开始时间")
    end_time = Column('end_time', VARCHAR(length=19), nullable=False,doc=u"生效结束时间")
    roster_type = Column('roster_type', SMALLINT(), nullable=False,doc=u"黑白名单类型")


class SlcMember(DeclarativeBase):
    """用户信息表"""
    __tablename__ = 'slc_member'

    __table_args__ = {}

    member_id = Column('member_id', INTEGER(), 
        Sequence('member_id_seq', start=100001, increment=1),
        primary_key=True,nullable=False,doc=u"用户id")
    node_id = Column('node_id', INTEGER(), nullable=False,doc=u"区域id")
    member_name = Column('member_name', VARCHAR(length=64), nullable=False,doc=u"用户登录名")
    password = Column('password', VARCHAR(length=128), nullable=False,doc=u"用户登录密码")
    realname = Column('realname', VARCHAR(length=64), nullable=False,doc=u"")
    idcard = Column('idcard', VARCHAR(length=32),doc=u"用户证件号码")
    sex = Column('sex', SMALLINT(), nullable=True,doc=u"用户性别0/1")
    age = Column('age', INTEGER(), nullable=True,doc=u"用户年龄")
    email = Column('email', VARCHAR(length=255), nullable=True,doc=u"用户邮箱")
    mobile = Column('mobile', VARCHAR(length=16), nullable=True,doc=u"用户手机")
    address = Column('address', VARCHAR(length=255), nullable=True,doc=u"用户地址")
    create_time = Column('create_time', VARCHAR(length=19), nullable=False,doc=u"创建时间")
    update_time = Column('update_time', VARCHAR(length=19), nullable=False,doc=u"更新时间")    
    

class SlcMemberOrder(DeclarativeBase):
    """
    订购信息表(交易记录)
    pay_status交易支付状态：0-未支付，1-已支付，2-已取消
    """
    __tablename__ = 'slc_member_order'

    __table_args__ = {}

    order_id = Column('order_id', VARCHAR(length=32),primary_key=True,nullable=False,doc=u"订单id")
    member_id = Column('member_id', INTEGER(),nullable=False,doc=u"用户id")
    product_id = Column('product_id', INTEGER(),nullable=False,doc=u"资费id")
    account_number = Column('account_number', VARCHAR(length=32),nullable=False,doc=u"上网账号")
    order_fee = Column('order_fee', INTEGER(), nullable=False,doc=u"订单费用")
    actual_fee = Column('actual_fee', INTEGER(), nullable=False,doc=u"实缴费用")
    pay_status = Column('pay_status', INTEGER(), nullable=False,doc=u"支付状态")
    accept_id = Column('accept_id', INTEGER(),nullable=False,doc=u"受理id")
    order_source = Column('order_source', VARCHAR(length=64), nullable=False,doc=u"订单来源")
    order_desc = Column('order_desc', VARCHAR(length=255),doc=u"订单描述")
    create_time = Column('create_time', VARCHAR(length=19), nullable=False,doc=u"交易时间")


class SlcRadAccount(DeclarativeBase):
    """
    上网账号表，每个会员可以同时拥有多个上网账号
    account_number 为每个套餐对应的上网账号，每个上网账号全局唯一
    用户状态 0:"预定",1:"正常", 2:"停机" , 3:"销户", 4:"到期"
    <radiusd default table>
    """

    __tablename__ = 'slc_rad_account'

    __table_args__ = {}

    account_number = Column('account_number', VARCHAR(length=32),primary_key=True,nullable=False,doc=u"上网账号")
    member_id = Column('member_id', INTEGER(),nullable=False,doc=u"用户id")
    product_id = Column('product_id', INTEGER(),nullable=False,doc=u"资费id")
    group_id = Column('group_id', INTEGER(),doc=u"用户组id")
    password = Column('password', VARCHAR(length=128), nullable=False,doc=u"上网密码")
    status = Column('status', INTEGER(), nullable=False,doc=u"用户状态")
    install_address = Column('install_address', VARCHAR(length=128), nullable=False,doc=u"装机地址")
    balance = Column('balance', INTEGER(), nullable=False,doc=u"用户余额-分")
    time_length = Column('time_length', INTEGER(), nullable=False,doc=u"用户时长-秒")
    expire_date = Column('expire_date', VARCHAR(length=10), nullable=False,doc=u"过期时间- ####-##-##")
    user_concur_number = Column('user_concur_number', INTEGER(), nullable=False,doc=u"用户并发数")
    bind_mac = Column('bind_mac', SMALLINT(), nullable=False,doc=u"是否绑定mac")
    bind_vlan = Column('bind_vlan', SMALLINT(), nullable=False,doc=u"是否绑定vlan")
    mac_addr = Column('mac_addr', VARCHAR(length=17),doc=u"mac地址")
    vlan_id = Column('vlan_id', INTEGER(),doc=u"内层vlan")
    vlan_id2 = Column('vlan_id2', INTEGER(),doc=u"外层vlan")
    ip_address = Column('ip_address', VARCHAR(length=15),doc=u"静态IP地址")
    last_pause = Column('last_pause', VARCHAR(length=19),doc=u"最后停机时间")
    create_time = Column('create_time', VARCHAR(length=19), nullable=False,doc=u"创建时间")
    update_time = Column('update_time', VARCHAR(length=19), nullable=False,doc=u"更新时间")

class SlcRadAccountAttr(DeclarativeBase):
    """上网账号扩展策略属性表"""
    __tablename__ = 'slc_rad_account_attr'
    __table_args__ = {}

    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"属性id")
    account_number = Column('account_number', VARCHAR(length=32),nullable=False,doc=u"上网账号")
    attr_name = Column(u'attr_name', VARCHAR(length=255), nullable=False,doc=u"属性名")
    attr_value = Column(u'attr_value', VARCHAR(length=255), nullable=False,doc=u"属性值")
    attr_desc = Column(u'attr_desc', VARCHAR(length=255),doc=u"属性描述")    

class SlcRadProduct(DeclarativeBase):
    '''资费信息表 <radiusd default table>'''
    __tablename__ = 'slc_rad_product'

    __table_args__ = {}

    id = Column('id', INTEGER(),primary_key=True,autoincrement=1,nullable=False,doc=u"资费id")
    node_id = Column('node_id', INTEGER(), nullable=False,doc=u"区域id")
    product_name = Column('product_name', VARCHAR(length=64), nullable=False,doc=u"资费名称")
    product_policy = Column('product_policy', INTEGER(), nullable=False,doc=u"资费策略")
    product_status = Column('product_status', SMALLINT(), nullable=False,doc=u"资费状态")    
    bind_mac = Column('bind_mac', SMALLINT(), nullable=False,doc=u"是否绑定mac")
    bind_vlan = Column('bind_vlan', SMALLINT(), nullable=False,doc=u"是否绑定vlan")
    concur_number = Column('concur_number', INTEGER(), nullable=False,doc=u"并发数")
    fee_period = Column('fee_period', VARCHAR(length=11),doc=u"开放认证时段")
    fee_months = Column('fee_months', INTEGER(),doc=u"买断月数")
    fee_price = Column('fee_price', INTEGER(), nullable=False,doc=u"资费价格")
    input_max_limit = Column('input_max_limit', INTEGER(), nullable=False,doc=u"上行速率")
    output_max_limit = Column('output_max_limit', INTEGER(), nullable=False,doc=u"下行速率")
    create_time = Column('create_time', VARCHAR(length=19), nullable=False,doc=u"创建时间")
    update_time = Column('update_time', VARCHAR(length=19), nullable=False,doc=u"更新时间")

class SlcRadProductAttr(DeclarativeBase):
    '''资费扩展属性表 <radiusd default table>'''
    __tablename__ = 'slc_rad_product_attr'

    __table_args__ = {}    

    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"属性id")
    product_id = Column('product_id', INTEGER(),nullable=False,doc=u"资费id")
    attr_name = Column(u'attr_name', VARCHAR(length=255), nullable=False,doc=u"属性名")
    attr_value = Column(u'attr_value', VARCHAR(length=255), nullable=False,doc=u"属性值")
    attr_desc = Column(u'attr_desc', VARCHAR(length=255),doc=u"属性描述")

class SlcRadBilling(DeclarativeBase):
    """计费信息表 is_deduct 0 未扣费 1 已扣费 <radiusd default table>"""
    __tablename__ = 'slc_rad_billing'

    __table_args__ = { }  

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"计费id")
    account_number = Column(u'account_number', VARCHAR(length=253), nullable=False,doc=u"上网账号")
    nas_addr = Column(u'nas_addr', VARCHAR(length=15), nullable=False,doc=u"bas地址")
    acct_session_id = Column(u'acct_session_id', VARCHAR(length=253), nullable=False,doc=u"会话id")
    acct_start_time = Column(u'acct_start_time', VARCHAR(length=19), nullable=False,doc=u"计费开始时间")
    acct_session_time = Column(u'acct_session_time', INTEGER(), nullable=False,doc=u"会话时长")
    acct_length = Column(u'acct_length', INTEGER(), nullable=False,doc=u"扣费时长")
    acct_fee = Column(u'acct_fee', INTEGER(), nullable=False,doc=u"应扣费用")
    actual_fee = Column('actual_fee', INTEGER(), nullable=False,doc=u"实扣费用")
    balance = Column('balance', INTEGER(), nullable=False,doc=u"当前余额")
    is_deduct = Column(u'is_deduct', INTEGER(), nullable=False,doc=u"是否扣费")
    create_time = Column('create_time', VARCHAR(length=19), nullable=False,doc=u"计费时间")


    
class SlcRadTicket(DeclarativeBase):
    """上网日志表 <radiusd default table>"""
    __tablename__ = 'slc_rad_ticket'

    __table_args__ = { }  

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"日志id")
    account_number = Column(u'account_number', VARCHAR(length=253), nullable=False,doc=u"上网账号")
    acct_input_gigawords = Column(u'acct_input_gigawords', INTEGER(),doc=u"会话的上行的字（4字节）的吉倍数")
    acct_output_gigawords = Column(u'acct_output_gigawords', INTEGER(),doc=u"会话的下行的字（4字节）的吉倍数")
    acct_input_octets = Column(u'acct_input_octets', INTEGER(),doc=u"会话的上行流量（字节数）")
    acct_output_octets = Column(u'acct_output_octets', INTEGER(),doc=u"会话的下行流量（字节数）")
    acct_input_packets = Column(u'acct_input_packets', INTEGER(),doc=u"会话的上行包数量")
    acct_output_packets = Column(u'acct_output_packets', INTEGER(),doc=u"会话的下行包数量")
    acct_session_id = Column(u'acct_session_id', VARCHAR(length=253), nullable=False,doc=u"会话id")
    acct_session_time = Column(u'acct_session_time', INTEGER(), nullable=False,doc=u"会话时长")
    acct_start_time = Column(u'acct_start_time', VARCHAR(length=19), nullable=False,doc=u"会话开始时间")
    acct_stop_time = Column(u'acct_stop_time', VARCHAR(length=19), nullable=False,doc=u"会话结束时间")
    acct_terminate_cause = Column(u'acct_terminate_cause',INTEGER(),doc=u"会话中止原因")
    mac_addr = Column(u'mac_addr', VARCHAR(length=128),doc=u"mac地址")
    calling_station_id =  Column(u'calling_station_id', VARCHAR(length=128),doc=u"用户接入物理信息")
    framed_netmask = Column(u'frame_id_netmask', VARCHAR(length=15),doc=u"地址掩码")
    framed_ipaddr = Column(u'framed_ipaddr', VARCHAR(length=15),doc=u"IP地址")
    nas_class = Column(u'nas_class', VARCHAR(length=253),doc=u"bas class")
    nas_addr = Column(u'nas_addr', VARCHAR(length=15), nullable=False,doc=u"bas地址")
    nas_port = Column(u'nas_port', VARCHAR(length=32),doc=u"接入端口")
    nas_port_id = Column(u'nas_port_id', VARCHAR(length=255),doc=u"接入端口物理信息")
    nas_port_type = Column(u'nas_port_type', INTEGER(),doc=u"接入端口类型")
    service_type = Column(u'service_type', INTEGER(),doc=u"接入服务类型")
    session_timeout = Column(u'session_timeout', INTEGER(),doc=u"会话超时时间")
    start_source = Column(u'start_source', INTEGER(), nullable=False,doc=u"会话开始来源")
    stop_source = Column(u'stop_source', INTEGER(), nullable=False,doc=u"会话中止来源")

    #relation definitions 

class SlcRadOnline(DeclarativeBase):
    """用户在线信息表 <radiusd default table>"""
    __tablename__ = 'slc_rad_online'

    __table_args__ = {
        'mysql_engine' : 'MEMORY'
    }  

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"在线id")
    account_number = Column(u'account_number', VARCHAR(length=32), nullable=False,doc=u"上网账号")
    nas_addr = Column(u'nas_addr', VARCHAR(length=32), nullable=False,doc=u"bas地址")
    acct_session_id = Column(u'acct_session_id', VARCHAR(length=64), nullable=False,doc=u"会话id")
    acct_start_time = Column(u'acct_start_time', VARCHAR(length=19), nullable=False,doc=u"会话开始时间")
    framed_ipaddr = Column(u'framed_ipaddr', VARCHAR(length=32), nullable=False,doc=u"IP地址")
    mac_addr = Column(u'mac_addr', VARCHAR(length=32), nullable=False,doc=u"mac地址")
    nas_port_id = Column(u'nas_port_id', VARCHAR(length=255), nullable=False,doc=u"接入端口物理信息")
    billing_times = Column(u'billing_times', INTEGER(), nullable=False,doc=u"已记账时间")
    start_source = Column(u'start_source', SMALLINT(), nullable=False,doc=u"会话开始来源")

class SlcRadAcceptLog(DeclarativeBase):
    '''
    业务受理日志表
    open:开户 pause:停机 resume:复机 cancel:销户 next:续费 charge:充值
    '''
    __tablename__ = 'slc_rad_accept_log'
    __table_args__ = {}

    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"日志id")
    accept_type = Column(u'accept_type', VARCHAR(length=16), nullable=False,doc=u"受理类型")
    accept_desc = Column(u'accept_desc', VARCHAR(length=512),doc=u"受理描述")
    account_number = Column(u'account_number', VARCHAR(length=32), nullable=False,doc=u"上网账号")
    operator_name = Column(u'operator_name', VARCHAR(32),doc=u"操作员名")
    accept_source = Column(u'accept_source', VARCHAR(length=128),doc=u"受理渠道来源")
    accept_time = Column(u'accept_time', VARCHAR(length=19), nullable=False,doc=u"受理时间")

class SlcRadOperateLog(DeclarativeBase):
    """操作日志表"""
    __tablename__ = 'slc_rad_operate_log'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False,doc=u"日志id")
    operator_name = Column(u'operator_name', VARCHAR(32), nullable=False,doc=u"操作员名称")
    operate_ip = Column(u'operate_ip', VARCHAR(length=128),doc=u"操作员ip")
    operate_time = Column(u'operate_time', VARCHAR(length=19), nullable=False,doc=u"操作时间")
    operate_desc = Column(u'operate_desc', VARCHAR(length=512),doc=u"操作描述")


def build_db(config=None):
    global engine
    engine = create_engine('mysql://%s:%s@%s:3306/mysql?charset=utf8'%(
                    config['user'],config['passwd'],config['host']))
    conn = engine.connect()
    try:
        conn.execute("drop database %s"%config['db'])
    except:
        pass
    conn.execute("create database %s DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci"%config['db'])
    conn.execute("commit")
    conn.close()
    engine = create_engine('mysql://%s:%s@%s:3306/%s?charset=utf8'%(
                    config['user'],config['passwd'],config['host'],config['db']))
    metadata.create_all(engine,checkfirst=True)    

def init_db(db):
    node = SlcNode()
    node.id = 1
    node.node_name = 'default'
    node.node_desc = u'测试区域'
    db.add(node)

    param1 = SlcParam()
    param1.param_name = 'max_session_timeout'
    param1.param_desc = u'最大会话时长(秒)'
    param1.param_value = '86400'
    db.add(param1)

    param2 = SlcParam()
    param2.param_name = 'reject_delay'
    param2.param_desc = u'拒绝延迟时间(秒),0-9'
    param2.param_value = '7'
    db.add(param2)
  

    opr = SlcOperator()
    opr.id = 1
    opr.node_id = 1
    opr.operator_name = 'admin'
    opr.operator_type = 1
    opr.operator_pass = md5('root').hexdigest()
    opr.operator_desc = 'admin'
    opr.operator_status = 1
    db.add(opr)

    bas = SlcRadBas()
    bas.id = 1
    bas.node_id = 1
    bas.vendor_id = '0'
    bas.ip_addr = '192.168.88.1'
    bas.bas_name = 'test_bas'
    bas.bas_secret = '123456'
    bas.coa_port = 3799
    bas.status = 1
    bas.time_type = 0
    db.add(bas)

    product = SlcRadProduct()
    product.id = 1
    product.node_id = 1
    product.product_name = u'10元包月套餐'
    product.product_policy = 0
    product.product_status = 1
    product.bind_mac = 0
    product.bind_vlan = 0
    product.concur_number = 0
    product.fee_num = 0
    product.fee_period = 0
    product.fee_price = 1000
    product.input_max_limit = 2097152
    product.output_max_limit = 2097152
    product.create_time = '2014-12-10 23:23:21'
    product.update_time = '2014-12-10 23:23:21'
    db.add(product)

    product2 = SlcRadProduct()
    product2.id = 2
    product2.node_id = 1
    product2.product_name = u'2元每小时'
    product2.product_policy = 1
    product2.product_status = 1
    product2.bind_mac = 0
    product2.bind_vlan = 0
    product2.concur_number = 0
    product2.fee_num = 0
    product2.fee_period = 0
    product2.fee_price = 200
    product2.input_max_limit = 2097152
    product2.output_max_limit = 2097152
    product2.create_time = '2014-12-10 23:23:21'
    product2.update_time = '2014-12-10 23:23:21'
    db.add(product2)


    member = SlcMember()
    member.member_id = 1000001
    member.member_name = 'tester'
    member.password = utils.encrypt('888888')
    member.node_id = 1
    member.realname = 'tester'
    member.idcard = '0'
    member.sex = '1'
    member.age = '33'
    member.email = '6583805@qq.com'
    member.mobile = '1366666666'
    member.address = 'hunan changsha'
    member.create_time = '2014-12-10 23:23:21'
    member.update_time = '2014-12-10 23:23:21'
    db.add(member)        
    account = SlcRadAccount()
    account.account_number = 'test01'
    account.member_id = member.member_id
    account.product_id = 1
    account.domain_name = 'cmcc'
    account.group_id = 1
    account.install_address = 'hunan'
    account.ip_address = ''
    account.mac_addr = ''
    account.password = utils.encrypt('888888')
    account.status = 1
    account.balance = 0
    account.basic_fee = 0
    account.time_length = 0
    account.flow_length = 0
    account.expire_date = '2015-12-30'
    account.user_concur_number = 0
    account.bind_mac = 0
    account.bind_vlan = 0
    account.vlan_id = 0
    account.vlan_id2 = 0
    account.create_time = '2014-12-10 23:23:21'
    account.update_time = '2014-12-10 23:23:21'
    db.add(account)

    account2 = SlcRadAccount()
    account2.account_number = 'test02'
    account2.member_id = member.member_id
    account2.product_id = 2
    account2.domain_name = 'cmcc'
    account2.group_id = 1
    account2.install_address = 'hunan'
    account2.ip_address = ''
    account2.mac_addr = ''
    account2.password = utils.encrypt('888888')
    account2.status = 1
    account2.balance = 1000
    account2.basic_fee = 0
    account2.time_length = 0
    account2.flow_length = 0
    account2.expire_date = '2015-12-30'
    account2.user_concur_number = 0
    account2.bind_mac = 0
    account2.bind_vlan = 0
    account2.vlan_id = 0
    account2.vlan_id2 = 0
    account2.create_time = '2014-12-10 23:23:21'
    account2.update_time = '2014-12-10 23:23:21'
    db.add(account2)    

    db.commit()

def init_test(db):
    for i in range(1000):
        member = SlcMember()
        member.member_id = 100000 + i
        member.member_name = 'tester%s'%i
        member.password = utils.encrypt('888888')
        member.node_id = 1
        member.realname = 'test00%s'%i
        member.idcard = '0'
        member.sex = '1'
        member.age = '33'
        member.email = 'wjt@lingyatech.com'
        member.mobile = '1366666666'
        member.address = 'hunan changsha'
        member.create_time = '2014-12-10 23:23:21'
        member.update_time = '2014-12-10 23:23:21'
        db.add(member)        
        account = SlcRadAccount()
        account.account_number = 'test00%s'%i
        account.member_id = member.member_id
        account.product_id = 1
        account.domain_name = 'cmcc'
        account.group_id = 1
        account.install_address = 'hunan'
        account.ip_address = ''
        account.mac_addr = ''
        account.password = utils.encrypt('888888')
        account.status = 1
        account.balance = 0
        account.basic_fee = 0
        account.time_length = 0
        account.flow_length = 0
        account.expire_date = '2015-12-30'
        account.user_concur_number = 0
        account.bind_mac = 0
        account.bind_vlan = 0
        account.vlan_id = 0
        account.vlan_id2 = 0
        account.create_time = '2014-12-10 23:23:21'
        account.update_time = '2014-12-10 23:23:21'
        db.add(account)
    db.commit()    



def install(config=None):

    print 'starting create and init database...'
    action = raw_input("drop and create database ?[n]")
    if action == 'y':
        build_db(config=config)

        db = scoped_session(sessionmaker(bind=engine, autocommit=False, autoflush=True))()  
        action = raw_input("init database ?[n]")
        if action == 'y':
            init_db(db)

        action = raw_input("init testdata ?[n]")
        if action == 'y':
            init_test(db)
            with open('./testusers.txt','wb') as tf:
                for i in range(1000):
                    tf.write('test00%s,%s\n'%(i,utils.encrypt('888888')))

def install2(config=None):
    print 'starting create and init database...'
    build_db(config=config)
    db = scoped_session(sessionmaker(bind=engine, autocommit=False, autoflush=True))()  
    init_db(db)

def update(config=None):
    print 'starting update database...'
    global engine
    engine = create_engine('mysql://%s:%s@%s:3306/%s?charset=utf8'%(
                    config['user'],config['passwd'],config['host'],config['db']))

    action = raw_input("rebuild database ?[n]")
    if action == 'y':
        metadata.drop_all(engine)      
    metadata.create_all(engine,checkfirst=True)    
    print 'starting update database done'
    db = scoped_session(sessionmaker(bind=engine, autocommit=False, autoflush=True))()  
    action = raw_input("init database ?[n]")
    if action == 'y':
        init_db(db)    



if __name__ == '__main__':
    install()

