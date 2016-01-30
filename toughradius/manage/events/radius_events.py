#!/usr/bin/env python
# coding=utf-8

from toughlib import  utils
from toughradius.manage import models
from toughradius.manage.settings import *
from toughlib.storage import Storage
import decimal
import datetime


class RadiusEvents:

    def __init__(self, dbengine=None,cache=None,**kwargs):
        self.dbengine = dbengine
        self.cache = cache

    def event_unlock_online(self, nasaddr, session_id):
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

def __call__(dbengine=None, mcache=None, **kwargs):
    return RadiusEvents(dbengine=dbengine, cache=mcache, **kwargs)

