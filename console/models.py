#!/usr/bin/env python
#coding:utf-8

from sqlalchemy import *
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relation

engine = create_engine('mysql://root:root@127.0.0.1:3306/slcrms?charset=utf8')
DeclarativeBase = declarative_base()
metadata = DeclarativeBase.metadata
metadata.bind = engine
 
class SlcNode(DeclarativeBase):
    __tablename__ = 'slc_node'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False)
    node_name = Column(u'node_name', VARCHAR(length=32), nullable=False)
    node_desc = Column(u'node_desc', VARCHAR(length=64), nullable=False)

    #relation definitions


class SlcRadBas(DeclarativeBase):
    __tablename__ = 'slc_rad_bas'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False)
    vendor_id = Column(u'vendor_id', VARCHAR(length=32), nullable=False)
    ip_addr = Column(u'ip_addr', VARCHAR(length=15), nullable=False)
    bas_name = Column(u'bas_name', VARCHAR(length=64), nullable=False)
    bas_secret = Column(u'secret', VARCHAR(length=32), nullable=False)
    status = Column(u'status', SMALLINT(), nullable=False)
    time_type = Column(u'time_type', SMALLINT(), nullable=False)

    #relation definitions


class SlcRadGroup(DeclarativeBase):
    __tablename__ = 'slc_rad_group'

    __table_args__ = {}

    #column definitions
    id = Column(u'id', INTEGER(), primary_key=True, nullable=False)
    group_name = Column(u'group_name', VARCHAR(length=64), nullable=False)
    group_desc = Column(u'group_desc', VARCHAR(length=255))
    bind_mac = Column(u'bind_mac', SMALLINT(), nullable=False)
    bind_vlan = Column(u'bind_vlan', SMALLINT(), nullable=False)
    concur_number = Column(u'concur_number', INTEGER(), nullable=False)
    update_time = Column(u'update_time', VARCHAR(length=19), nullable=False)

    #relation definitions


class SlcRadRoster(DeclarativeBase):
    """黑白名单
    """
    __tablename__ = 'slc_rad_roster'

    __table_args__ = {}

    #column definitions
    account_number = Column('account_number', VARCHAR(length=32))
    begin_time = Column('begin_time', VARCHAR(length=19), nullable=False)
    end_time = Column('end_time', VARCHAR(length=19), nullable=False)
    mac_addr = Column('mac_addr', VARCHAR(length=17), primary_key=True, nullable=False)
    roster_type = Column('roster_type', SMALLINT(), nullable=False)


class SlcRadAcct(DeclarativeBase):
    """
    Radius扣费信息
    """
    __tablename__ = 'slc_rad_acct'

    __table_args__ = {}

    acct_id = Column('acct_id', INTEGER(), primary_key=True,autoincrement=1,nullable=False)
    member_id = Column('member_id', INTEGER(),nullable=False)
    account_number = Column('account_number', VARCHAR(length=32), nullable=False)
    acct_fee = Column('acct_fee', INTEGER(), nullable=False)
    acct_session_time = Column('acct_session_time', INTEGER(), nullable=False)
    acct_start_time = Column('acct_start_time', VARCHAR(length=19), nullable=False)
    acct_stop_time = Column('acct_stop_time', VARCHAR(length=19), nullable=False)
    actual_fee = Column('actual_fee', INTEGER(), nullable=False)
    is_deduct = Column('is_deduct', SMALLINT(), nullable=False)
    operate_time = Column('operate_time', VARCHAR(length=19), nullable=False)


class SlcMember(DeclarativeBase):
    """
    会员信息
    """
    __tablename__ = 'slc_member'

    __table_args__ = {}

    member_id = Column('member_id', INTEGER(), 
        Sequence('member_id_seq', start=100001, increment=1),
        primary_key=True,nullable=False)
    node_id = Column('node_id', INTEGER(), nullable=False)
    member_name = Column('member_name', VARCHAR(length=32), nullable=False)
    realname = Column('realname', VARCHAR(length=64), nullable=False)
    idcard = Column('idcard', VARCHAR(length=32))
    sex = Column('sex', SMALLINT(), nullable=True)
    age = Column('age', INTEGER(), nullable=True)
    email = Column('email', VARCHAR(length=255), nullable=True)
    mobile = Column('mobile', VARCHAR(length=16), nullable=True)
    address = Column('address', VARCHAR(length=255), nullable=True)
    create_time = Column('create_time', VARCHAR(length=19), nullable=False)
    update_time = Column('update_time', VARCHAR(length=19), nullable=False)    
    

class SlcMemberOrder(DeclarativeBase):
    """
    会员订购信息(交易记录)
    pay_status交易支付状态：0-未支付，1-已支付，2-已取消
    """
    __tablename__ = 'slc_member_order'

    __table_args__ = {}

    order_id = Column('order_id', VARCHAR(length=32),primary_key=True,nullable=False)
    member_id = Column('member_id', INTEGER(),nullable=False)
    product_id = Column('product_id', INTEGER(),nullable=False)
    account_number = Column('account_number', VARCHAR(length=32),nullable=False)
    order_fee = Column('order_fee', INTEGER(), nullable=False)
    pay_status = Column('pay_status', INTEGER(), nullable=False)
    order_detail = Column('order_detail', VARCHAR(length=1024), nullable=False)
    create_time = Column('create_time', VARCHAR(length=19), nullable=False)


class SlcMemberRefund(DeclarativeBase):
    """
    会员退款记录
    status退款状态：0-未完成，1-已完成，2-已取消
    """
    __tablename__ = 'slc_member_refund'

    __table_args__ = {}

    refund_id = Column('refund_id', VARCHAR(length=32),primary_key=True,nullable=False)
    member_id = Column('member_id', INTEGER(),nullable=False)
    product_id = Column('product_id', INTEGER(),nullable=False)
    account_number = Column('account_number', VARCHAR(length=32),nullable=False)
    refund_fee = Column('refund_fee', INTEGER(), nullable=False)
    status = Column('status', INTEGER(), nullable=False)
    order_detail = Column('order_detail', VARCHAR(length=1024), nullable=False)
    create_time = Column('create_time', VARCHAR(length=19), nullable=False)


class SlcRadAccount(DeclarativeBase):
    """
    会员上网账号，每个会员可以同时拥有多个上网账号
    account_number 为每个套餐对应的上网账号，每个上网账号全局唯一
    """

    __tablename__ = 'slc_rad_account'

    __table_args__ = {}

    account_number = Column('account_number', VARCHAR(length=32),primary_key=True,nullable=False)
    member_id = Column('member_id', INTEGER(),nullable=False)
    product_id = Column('product_id', INTEGER(),nullable=False)
    group_id = Column('group_id', VARCHAR(length=32))
    password = Column('password', VARCHAR(length=64), nullable=False)
    status = Column('status', INTEGER(), nullable=False)
    install_address = Column('install_address', VARCHAR(length=128), nullable=False)
    balance = Column('balance', INTEGER(), nullable=False)
    basic_fee = Column('basic_fee', INTEGER(), nullable=False)
    time_length = Column('time_length', INTEGER(), nullable=False)
    flow_length = Column('flow_length', INTEGER(), nullable=False)
    expire_date = Column('expire_date', VARCHAR(length=10), nullable=False)
    user_concur_number = Column('user_concur_number', INTEGER(), nullable=False)
    bind_mac = Column('user_mac', SMALLINT(), nullable=False)
    bind_vlan = Column('user_vlan', SMALLINT(), nullable=False)
    mac_addr = Column('mac_addr', VARCHAR(length=17))
    vlan_id = Column('vlan_id', INTEGER())
    vlan_id2 = Column('vlan_id2', INTEGER())
    ip_address = Column('ip_address', VARCHAR(length=15))
    domain_name = Column('domain_name', VARCHAR(length=32))
    create_time = Column('create_time', VARCHAR(length=19), nullable=False)
    update_time = Column('update_time', VARCHAR(length=19), nullable=False)

class SlcRadProduct(DeclarativeBase):
    '''宽带产品信息
    '''
    __tablename__ = 'slc_rad_product'

    __table_args__ = {}

    product_id = Column('product_id', INTEGER(),primary_key=True,autoincrement=1,nullable=False)
    product_name = Column('product_name', VARCHAR(length=64), nullable=False)
    product_policy = Column('product_policy', INTEGER(), nullable=False)
    product_status = Column('product_status', SMALLINT(), nullable=False)    
    domain_name = Column('domain_name', VARCHAR(length=32))
    bandwidth_code = Column('bandwidth_code', VARCHAR(length=32), nullable=False)
    bind_mac = Column('bind_mac', SMALLINT(), nullable=False)
    bind_vlan = Column('bind_vlan', SMALLINT(), nullable=False)
    concur_number = Column('concur_number', INTEGER(), nullable=False)
    fee_num = Column('fee_num', INTEGER(), nullable=False) 
    fee_period = Column('fee_period', VARCHAR(length=11))
    fee_price = Column('fee_price', INTEGER(), nullable=False)
    input_max_limit = Column('input_max_limit', INTEGER(), nullable=False)
    output_max_limit = Column('output_max_limit', INTEGER(), nullable=False)
    input_rate_code = Column('input_rate_code', VARCHAR(length=32), nullable=False)
    output_rate_code = Column('output_rate_code', VARCHAR(length=32), nullable=False)
    create_time = Column('create_time', VARCHAR(length=19), nullable=False)
    update_time = Column('update_time', VARCHAR(length=19), nullable=False)


def build_db():
    metadata.create_all(engine,checkfirst=True)

def rebuild_db():
    metadata.drop_all(engine)
    metadata.create_all(engine,checkfirst=True)    

def init_db():
    from sqlalchemy.orm import scoped_session, sessionmaker
    db = scoped_session(sessionmaker(bind=engine, autocommit=False, autoflush=True))()

    node = SlcNode()
    node.id = 1
    node.node_name = 'default'
    node.node_desc = 'default'
    db.add(node)

    bas = SlcRadBas()
    bas.id = 1
    bas.vendor_id = '14896'
    bas.ip_addr = '127.0.0.1'
    bas.bas_name = 'test_bas'
    bas.bas_secret = '123456'
    bas.status = 1
    bas.time_type = 0
    db.add(bas)

    product = SlcRadProduct()
    product.product_id = 1
    product.product_name = 'test_product'
    product.product_policy = 0
    product.product_status = 1
    product.domain_name = 'cmcc'
    product.bandwidth_code = 'none'
    product.bind_mac = 0
    product.bind_vlan = 0
    product.concur_number = 0
    product.fee_num = 0
    product.fee_period = 0
    product.fee_price = 0
    product.input_max_limit = 102400
    product.output_max_limit = 102400
    product.input_rate_code = 'none'
    product.output_rate_code = 'none'
    product.create_time = '2014-12-10 23:23:21'
    product.update_time = '2014-12-10 23:23:21'
    db.add(product)

    member = SlcMember()
    member.member_id = 100001
    member.node_id = 1
    member.member_name = 'wjt'
    member.realname = 'wjt'
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
    account.account_number = 'wjt001'
    account.member_id = 100001
    account.product_id = 1
    account.domain_name = 'cmcc'
    account.group_id = 1
    account.install_address = 'hunan'
    account.ip_address = ''
    account.mac_addr = ''
    account.password = '888888'
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

if __name__ == '__main__':
    action = raw_input("is rebuild?[n]")
    if action == 'y':
        rebuild_db()
    else:
        build_db()

    action = raw_input("init_db ?[n]")
    if action == 'y':
        init_db()



