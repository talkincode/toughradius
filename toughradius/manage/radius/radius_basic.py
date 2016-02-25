#!/usr/bin/env python
# coding=utf-8

from toughlib import  utils
from toughradius.manage import models
from toughradius.manage.settings import *
from toughlib.storage import Storage
from toughlib.utils import timecast
import decimal
import datetime
import copy
decimal.getcontext().prec = 16
decimal.getcontext().rounding = decimal.ROUND_UP

class RadiusBasic:

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        self.dbengine = dbengine
        self.cache = cache
        self.aes = aes
        self.request = Storage(request)
        self.account = self.get_account_by_username(self.request.account_number)

    #@timecast
    def get_param_value(self, name, defval=None):
        def fetch_result():
            table = models.TrParam.__table__
            with self.dbengine.begin() as conn:
                return conn.execute(
                    table.select().with_only_columns([table.c.param_value]).where(
                        table.c.param_name==name)).scalar() or defval
        return self.cache.aget(param_cache_key(name),fetch_result, expire=600)

    #@timecast
    def get_account_by_username(self,username):
        def fetch_result():
            table = models.TrAccount.__table__
            with self.dbengine.begin() as conn:
                val = conn.execute(table.select().where(
                    table.c.account_number==username)).first()
                return val and Storage(val.items()) or None
        return self.cache.aget(account_cache_key(username),fetch_result, expire=600)

    #@timecast
    def get_product_by_id(self,product_id):
        def fetch_result():
            table = models.TrProduct.__table__
            with self.dbengine.begin() as conn:
                val = conn.execute(table.select().where(table.c.id==product_id)).first()
                return val and Storage(val.items()) or None
        return self.cache.aget(product_cache_key(product_id),fetch_result,expire=600)

    #@timecast
    def get_product_attrs(self,product_id):
        def fetch_result():
            table = models.TrProductAttr.__table__
            with self.dbengine.begin() as conn:
                vals = conn.execute(table.select().where(
                    table.c.product_id==product_id).where(
                    table.c.attr_type==1))
                return vals and [Storage(val.items()) for val in vals] or []
        return self.cache.aget(product_attrs_cache_key(product_id),fetch_result,expire=600)


    def get_user_balance(self):
        table = models.TrAccount.__table__
        with self.dbengine.begin() as conn:
            return conn.execute(
                table.select().with_only_columns([table.c.balance]).where(
                    table.c.account_number==self.account.account_number)).scalar()

    def get_user_time_length(self):
        table = models.TrAccount.__table__
        with self.dbengine.begin() as conn:
            return conn.execute(
                table.select(table.c.time_length).with_only_columns([table.c.time_length]).where(
                    table.c.account_number==self.account.account_number)).scalar()

    def get_user_flow_length(self):
        table = models.TrAccount.__table__
        with self.dbengine.begin() as conn:
            return conn.execute(
                table.select(table.c.flow_length).with_only_columns([table.c.flow_length]).where(
                    table.c.account_number==self.account.account_number)).scalar()

    def update_user_mac(self, macaddr):
        table = models.TrAccount.__table__
        with self.dbengine.begin() as conn:
            stmt = table.update().where(
                table.c.account_number==self.account.account_number).values(mac_addr=macaddr)
            conn.execute(stmt)

    def update_user_vlan_id1(self, vlan_id1):
        table = models.TrAccount.__table__
        with self.dbengine.begin() as conn:
            stmt = table.update().where(
                table.c.account_number==self.account.account_number).values(vlan_id1=vlan_id1)
            conn.execute(stmt)

    def update_user_vlan_id2(self, vlan_id2):
        table = models.TrAccount.__table__
        with self.dbengine.begin() as conn:
            stmt = table.update().where(
                table.c.account_number==self.account.account_number).values(vlan_id2=vlan_id2)
            conn.execute(stmt)

    def get_online(self, nasaddr, session_id):
        table = models.TrOnline.__table__
        with self.dbengine.begin() as conn:
            return conn.execute(table.select().where(
                table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id)).first()

    def add_online(self,online):
        table = models.TrOnline.__table__
        with self.dbengine.begin() as conn:
            conn.execute(table.insert().values(**online))

    def is_online(self, nasaddr, session_id):
        table = models.TrOnline.__table__
        with self.dbengine.begin() as conn:
            return conn.execute(table.count().where(
                table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id)).scalar() > 0

    def del_online(self, nasaddr, session_id):
        table = models.TrOnline.__table__
        with self.dbengine.begin() as conn:
            stmt = table.delete().where(
                table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id)
            conn.execute(stmt)

    def count_online(self):
        table = models.TrOnline.__table__
        with self.dbengine.begin() as conn:
            return conn.execute(table.count().where(
                table.c.account_number==self.account.account_number)).scalar()

    def update_online(self, nasaddr, session_id, **kwargs):
        table = models.TrOnline.__table__
        with self.dbengine.begin() as conn:
            stmt = table.update().where(table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id).values(**kwargs)
            conn.execute(stmt)

    def disconnect(self,nasaddr, session_id):
        pass


    def get_input_total(self):
        bl = decimal.Decimal(self.request.acct_input_octets)/decimal.Decimal(1024)
        gl = decimal.Decimal(self.request.acct_input_gigawords)*decimal.Decimal(4*1024*1024)
        tl = bl + gl
        return int(tl.to_integral_value())

    def get_output_total(self):
        bl = decimal.Decimal(self.request.acct_output_octets)/decimal.Decimal(1024)
        gl = decimal.Decimal(self.request.acct_output_gigawords)*decimal.Decimal(4*1024*1024)
        tl = bl + gl
        return int(tl.to_integral_value())

    def add_ticket(self,ticket):
        table = models.TrTicket.__table__
        data = {k.name:ticket[k.name] for k in table.columns if k.name not in 'id'}
        with self.dbengine.begin() as conn:
            conn.execute(table.insert().values(**data))

    def update_billing(self, billing):
        acctount_table = models.TrAccount.__table__
        bill_table = models.TrBilling.__table__
        online_table = models.TrOnline.__table__

        with self.dbengine.begin() as conn:
            conn.execute(acctount_table.update().where(
                acctount_table.c.account_number==billing.account_number).values(
                    balance=billing.balance,
                    time_length=billing.time_length,
                    flow_length=billing.flow_length))

            conn.execute(bill_table.insert().values(**billing))

            conn.execute(online_table.update().where(
                online_table.c.nas_addr==billing.nas_addr).where(
                    acct_session_id==billing.acct_session_id).values(
                        billing_times=billing.billing_times,
                        input_total=billing.input_total,
                        output_total=billing.output_total))

    def unlock_online(self, nasaddr, session_id):
        online_table = models.TrOnline.__table__
        ticket_table = models.TrTicket.__table__
        def new_ticket(online):
            _datetime = datetime.datetime.now()
            _starttime = datetime.datetime.strptime(online.acct_start_time,"%Y-%m-%d %H:%M:%S")
            session_time = (_datetime - _starttime).seconds
            stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
            ticket = Storage()
            ticket.account_number = online.account_number,
            ticket.acct_session_id = online.acct_session_id,
            ticket.acct_start_time = online.acct_start_time,
            ticket.nas_addr = online.nas_addr,
            ticket.framed_ipaddr = online.framed_ipaddr,
            ticket.acct_session_time = session_time,
            ticket.acct_stop_time = stop_time,
            return ticket

        if all((nasaddr,session_id)):
            with self.dbengine.begin() as conn:
                online = conn.execute(online_table.select().where(
                    online_table.c.nas_addr==nasaddr).where(
                    acct_session_id==session_id)).first()

                ticket = new_ticket(online)
                conn.execute(ticket_table.insert().values(**ticket))

                conn.execute(online_table.delete().where(
                    online_table.c.nas_addr==nasaddr).where(
                    acct_session_id==session_id))

        elif nasaddr and not session_id:
            with self.dbengine.begin() as conn:
                onlines = conn.execute(online_table.select().where(online_table.c.nas_addr==nasaddr))
                tickets = (new_ticket(ol) for ol in onlines)
                with self.dbengine.begin() as conn:
                    conn.execute(ticket_table.insert(),tickets)
                    conn.execute(online_table.delete().where(online_table.c.nas_addr==nasaddr))









