#!/usr/bin/env python
# coding=utf-8

from toughlib import  utils
from toughradius.manage import models
from toughradius.manage.settings import *
from toughlib.storage import Storage
from toughlib.utils import timecast
import decimal

decimal.getcontext().prec = 16
decimal.getcontext().rounding = decimal.ROUND_UP

class RadiusBasic:

    def __init__(self, app, request):
        self.app = app
        self.cache = self.app.mcache
        self.request = request
        self.log = self.app.syslog
        self.account = self.get_account_by_username(self.request['username'])
        self._ticket = None
        
    #@timecast
    def get_param_value(self, name, defval=None):
        def fetch_result():
            table = models.TrParam.__table__
            with self.app.db_engine.begin() as conn:
                return conn.execute(
                    table.select().with_only_columns([table.c.param_value]).where(
                        table.c.param_name==name)).scalar() or defval
        return self.cache.aget(param_cache_key(name),fetch_result, expire=600)

    #@timecast
    def get_account_by_username(self,username):
        def fetch_result():
            table = models.TrAccount.__table__
            with self.app.db_engine.begin() as conn:
                val = conn.execute(table.select().where(
                    table.c.account_number==username)).first()
                return val and Storage(val.items()) or None
        return self.cache.aget(account_cache_key(username),fetch_result, expire=600)

    #@timecast
    def get_product_by_id(self,product_id):
        def fetch_result():
            table = models.TrProduct.__table__
            with self.app.db_engine.begin() as conn:
                val = conn.execute(table.select().where(table.c.id==product_id)).first()
                return val and Storage(val.items()) or None
        return self.cache.aget(product_cache_key(product_id),fetch_result,expire=600)

    #@timecast
    def get_product_attrs(self,product_id):
        def fetch_result():
            table = models.TrProductAttr.__table__
            with self.app.db_engine.begin() as conn:
                vals = conn.execute(table.select().where(
                    table.c.product_id==product_id).where(
                    table.c.attr_type==1))
                return vals and [Storage(val.items()) for val in vals] or []
        return self.cache.aget(product_attrs_cache_key(product_id),fetch_result,expire=600)


    def get_user_balance(self):
        table = models.TrAccount.__table__
        with self.app.db_engine.begin() as conn:
            return conn.execute(
                table.select().with_only_columns([table.c.balance]).where(
                    table.c.account_number==self.account.account_number)).scalar()

    def get_user_time_length(self):
        table = models.TrAccount.__table__
        with self.app.db_engine.begin() as conn:
            return conn.execute(
                table.select(table.c.time_length).with_only_columns([table.c.time_length]).where(
                    table.c.account_number==self.account.account_number)).scalar()        

    def get_user_flow_length(self):
        table = models.TrAccount.__table__
        with self.app.db_engine.begin() as conn:
            return conn.execute(
                table.select(table.c.flow_length).with_only_columns([table.c.flow_length]).where(
                    table.c.account_number==self.account.account_number)).scalar()           

    def update_user_mac(self, macaddr):
        table = models.TrAccount.__table__
        with self.app.db_engine.begin() as conn:
            stmt = table.update().where(
                table.c.account_number==self.account.account_number).values(mac_addr=macaddr)
            conn.execute(stmt)

    def update_user_vlan_id1(self, vlan_id1):
        table = models.TrAccount.__table__
        with self.app.db_engine.begin() as conn:
            stmt = table.update().where(
                table.c.account_number==self.account.account_number).values(vlan_id1=vlan_id1)
            conn.execute(stmt)

    def update_user_vlan_id2(self, vlan_id2):
        table = models.TrAccount.__table__
        with self.app.db_engine.begin() as conn:
            stmt = table.update().where(
                table.c.account_number==self.account.account_number).values(vlan_id2=vlan_id2)
            conn.execute(stmt)

    def get_online(self, nasaddr, session_id):
        table = models.TrOnline.__table__
        with self.app.db_engine.begin() as conn:
            return conn.execute(table.select().where(
                table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id)).first()        

    def is_online(self, nasaddr, session_id):
        table = models.TrOnline.__table__
        with self.app.db_engine.begin() as conn:
            return conn.execute(table.count().where(
                table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id)) > 0

    def del_online(self, nasaddr, session_id):
        table = models.TrOnline.__table__
        with self.app.db_engine.begin() as conn:
            stmt = table.delete().where(
                table.c.nas_addr==nasaddr).where(
                table.c.acct_session_id==session_id)
            conn.execute(stmt)

    def count_online(self):
        table = models.TrOnline.__table__
        with self.app.db_engine.begin() as conn:
            return conn.execute(table.count().where(
                table.c.account_number==self.account.account_number))

    def update_online(self, nasaddr, session_id, **kwargs):
        table = models.TrOnline.__table__
        with self.app.db_engine.begin() as conn:
            stmt = table.update().where(table.c.nas_addr==nasaddr).where(
                acct_session_id==session_id).values(**kwargs)
            conn.execute(stmt)

    def disconnect(self,nasaddr, session_id):
        pass


    def get_input_total(self):
        bl = decimal.Decimal(self.ticket.acct_input_octets)/decimal.Decimal(1024)
        gl = decimal.Decimal(self.ticket.acct_input_gigawords)*decimal.Decimal(4*1024*1024)
        tl = bl + gl
        return int(tl.to_integral_value())   
        
    def get_output_total(self):
        bl = decimal.Decimal(self.ticket.acct_input_octets)/decimal.Decimal(1024)
        gl = decimal.Decimal(self.ticket.acct_output_gigawords)*decimal.Decimal(4*1024*1024)
        tl = bl + gl
        return int(tl.to_integral_value())      

    @property
    def ticket(self):
        if self._ticket:
            return self._ticket
        else:
            self._ticket = Storage(
                account_number = self.request['username'],
                mac_addr = self.request['macaddr'],
                nas_addr = self.request['nasaddr'],
                nas_port = self.request['nas_port'],
                service_type = self.request.get('service_type','radius'),
                framed_ipaddr = self.request['ipaddr'],
                framed_netmask = self.request.get('netmask',''),
                nas_class = self.request.get('nas_class',''),
                session_timeout = self.request['session_timeout'],
                calling_stationid = self.request.get('calling_stationid',''),
                acct_status_type = self.request['acct_status_type'],
                acct_input_octets = self.request['input_octets'],
                acct_output_octets = self.request['output_octets'],
                acct_session_id = self.request['session_id'],
                acct_session_time = self.request['session_time'],
                acct_input_packets = self.request['input_pkts'],
                acct_output_packets = self.request['output_pkts'],
                acct_terminate_cause = self.request['terminate_cause'],
                acct_input_gigawords = self.request.get('input_gigawords',0),
                acct_output_gigawords = self.request.get('output_gigawords',0),
                event_timestamp = self.request['event_timestamp'],
                nas_port_type=self.grequest['nas_port_type'],
                nas_port_id=self.request['nas_port_id']
            )
        return self._ticket

    def add_ticket(self,ticket):
        table = models.TrTicket.__table__
        with self.app.db_engine.begin() as conn:
            conn.execute(table.insert().values(**ticket))


    def update_billing(self, billing):
        acctount_table = models.TrAccount.__table__
        bill_table = models.TrBilling.__table__
        online_table = models.TrOnline.__table__

        with self.app.db_engine.begin() as conn:
            conn.execute(acctount_table.update().where(
                acctount_table.c.account_number==billing.account_number).values(
                    balance=balance, time_length=time_length, flow_length=flow_length))

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
            ticket = models.TrTicket()
            ticket.account_number = online.account_number,
            ticket.acct_session_id = online.acct_session_id,
            ticket.acct_start_time = online.acct_start_time,
            ticket.nas_addr = online.nas_addr,
            ticket.framed_ipaddr = online.framed_ipaddr,
            ticket.acct_session_time = session_time,
            ticket.acct_stop_time = stop_time,
            return ticket

        if all((nasaddr,session_id)):
            with self.app.db_engine.begin() as conn:
                online = conn.execute(online_table.select().where(
                    online_table.c.nas_addr==nasaddr).where(
                    acct_session_id==session_id)).first()

                ticket = new_ticket(online)
                conn.execute(ticket_table.insert().values(**ticket))
                
                conn.execute(online_table.delete().where(
                    online_table.c.nas_addr==nasaddr).where(
                    acct_session_id==session_id))

        elif nas_addr and not session_id:
            onlines = conn.execute(online_table.select().where(online_table.c.nas_addr==nasaddr))
            tickets = (new_ticket(online) for ol in onlines)
            with self.app.db_engine.begin() as conn:
                conn.execute(ticket_table.insert(),tickets)
                conn.execute(online_table.delete().where(online_table.c.nas_addr==nasaddr))









