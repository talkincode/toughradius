#!/usr/bin/env python
#coding=utf-8
import MySQLdb
from MySQLdb import cursors
from DBUtils.PooledDB import PooledDB
import settings
import cache
import datetime

ticket_fds = [
    'account_number','acct_fee','acct_input_gigawords','acct_input_octets',
    'acct_input_packets','acct_output_gigawords','acct_output_octets',
    'acct_output_packets','acct_session_id','acct_session_time',
    'acct_start_time','acct_stop_time','acct_terminate_cause',
    'calling_station_id','fee_receivables','frame_netmask',
    'framed_ipaddr','is_deduct','nas_class','nas_addr',
    'nas_port','nas_port_id','nas_port_type','service_type',
    'session_timeout','start_source','stop_source',"mac_addr"
]

get_cursor = lambda conn: conn.cursor(cursors.DictCursor)

class Connect:
    def __init__(self, dbpool):
        self.conn = dbpool.connect()

    def __enter__(self):
        return self.conn   

    def __exit__(self, exc_type, exc_value, exc_tb):
        self.conn.close()

class Cursor:
    def __init__(self, dbpool):
        self.conn = dbpool.connect()
        self.cursor = get_cursor(self.conn)

    def __enter__(self):
        return self.cursor 

    def __exit__(self, exc_type, exc_value, exc_tb):
        self.conn.close()

class DbPool():
    dbpool = None
    def connect(self):
        if not self.dbpool:
            self.dbpool = PooledDB(creator=MySQLdb,**settings.db_config['mysql'])
        return self.dbpool.connection()

class Store():

    def __init__(self,dbpool=None):
        self.dbpool = dbpool 

    @cache.cache('all')
    def get_param(self,param_name):
        with Cursor(self.dbpool) as cur:
            cur.execute("select param_value from  slc_param where param_name = %s",(param_name,))
            param = cur.fetchone()
            return param and param['param_value'] or None

    def update_param_cache(self):
        with Cursor(self.dbpool) as cur:
            cur.execute("select param_name from  slc_param where param_name = %s",(param_name,))
            for param in cur:
                cache.delete('all',self.get_param, param['param_name'])

    def list_bas(self):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from  slc_rad_bas")
            return [bas for bas in cur] 

    @cache.cache('all')
    def get_user(self,username):
        with Cursor(self.dbpool) as cur:
            cur.execute("select a.*,p.product_policy from slc_rad_account a,slc_rad_product p "
                "where a.product_id = p.id and a.account_number = %s ",(username,))
            user =  cur.fetchone()
            return user

    @cache.cache('all')
    def get_user_attrs(self,username):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_account_attr where account_number = %s ",(username,))
            return cur.fetchall()  

    def update_user_cache(self,username):
        cache.delete('all',self.get_user, username)
        cache.delete('all',self.get_user_attrs, username)

    def update_user_balance(self,username,balance):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set balance = %s where account_number = %s"
            cur.execute(sql,(balance,username))
            conn.commit()
            self.update_user_cache(username)

    def update_user_mac(self,username,mac_addr):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set mac_addr = %s where account_number = %s"
            cur.execute(sql,(mac_addr,username))
            conn.commit()
            self.update_user_cache(username)

    def update_user_vlan_id(self,username,vlan_id):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set vlan_id = %s where account_number = %s"
            cur.execute(sql,(vlan_id,username))
            conn.commit()
            self.update_user_cache(username)

    def update_user_vlan_id2(self,username,vlan_id2):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set vlan_id2 = %s where account_number = %s"
            cur.execute(sql,(vlan_id2,username))    
            conn.commit() 
            self.update_user_cache(username)     

    @cache.cache('all')
    def get_group(self,group_id):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_group where id = %s ",(group_id,))
            return cur.fetchone()

    def clear_group_cache(self,group_id):
        cache.delete('all',self.get_group, group_id)

    @cache.cache('all')
    def get_product(self,product_id):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_product where id = %s ",(product_id,))
            return cur.fetchone()  

    @cache.cache('all')
    def get_product_attrs(self,product_id):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_product_attr where product_id = %s ",(product_id,))
            return cur.fetchall()  

    def clear_product_cache(self,product_id):
        cache.delete('all',self.get_product, product_id)
        cache.delete('all',self.get_product_attrs, product_id)
          
    def is_online(self,nas_addr,acct_session_id):
        with Cursor(self.dbpool) as cur: 
            sql = 'select count(id) as online from slc_rad_online where  nas_addr = %s and acct_session_id = %s'
            cur.execute(sql,(nas_addr,acct_session_id)) 
            return cur.fetchone()['online'] > 0

    def get_online(self,nas_addr,acct_session_id):
        with Cursor(self.dbpool) as cur: 
            sql = 'select * from slc_rad_online where  nas_addr = %s and acct_session_id = %s'
            cur.execute(sql,(nas_addr,acct_session_id)) 
            return cur.fetchone()     

    def get_nas_onlines(self,nas_addr):
        with Cursor(self.dbpool) as cur: 
            sql = 'select * from slc_rad_online where nas_addr = %s'
            cur.execute(sql,(nas_addr,)) 
            return cur.fetchall()        

    def add_online(self,online):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            keys = ','.join(online.keys())
            vals = ",".join(["'%s'"% c for c in online.values()])
            sql = 'insert into slc_rad_online (%s) values(%s)'%(keys,vals)
            cur.execute(sql)
            conn.commit()
    
    def del_online(self,nas_addr,acct_session_id):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = 'delete from slc_rad_online where nas_addr = %s and acct_session_id = %s'
            cur.execute(sql,(nas_addr,acct_session_id))
            conn.commit()

    # def del_nas_onlines(self,nas_addr):
    #     with Connect(self.dbpool) as conn:
    #         cur = conn.cursor()
    #         sql = 'delete from slc_rad_online where nas_addr = %s'
    #         cur.execute(sql,(nas_addr,))
    #         conn.commit()

    def add_ticket(self,ticket):
        _ticket = ticket.copy()
        for _key in _ticket:
            if _key not in ticket_fds:
                del ticket[_key]
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            keys = ','.join(ticket.keys())
            vals = ",".join(["'%s'"% c for c in ticket.values()])
            sql = 'insert into slc_rad_ticket (%s) values(%s)'%(keys,vals)
            cur.execute(sql)
            conn.commit()

    def unlock_online(self,nas_addr,acct_session_id,stop_source):
        bsql = ''' insert into slc_rad_ticket (
        account_number,acct_session_id,acct_start_time,nas_addr,framed_ipaddr,start_source,
        acct_session_time,acct_stop_time,stop_source,fee_receivables,acct_fee,is_deduct) values(
         %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ''' 
        def _ticket(online):
            ticket = []
            ticket.append(online['account_number'])
            ticket.append(online['acct_session_id'])
            ticket.append(online['acct_start_time'])
            ticket.append(online['nas_addr'])
            ticket.append(online['framed_ipaddr'])
            ticket.append(online['start_source'])
            _datetime = datetime.datetime.now()
            _starttime = datetime.datetime.strptime(online['acct_start_time'],"%Y-%m-%d %H:%M:%S")
            session_time = (_datetime - _starttime).seconds
            stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
            ticket.append(session_time)
            ticket.append(stop_time)
            ticket.append(stop_source)
            ticket.append(0)
            ticket.append(0)
            ticket.append(0)
            return ticket

        def _unlock_one():
            ticket = None
            with Connect(self.dbpool) as conn:
                cur = conn.cursor(cursors.DictCursor)
                sql = 'select * from slc_rad_online where  nas_addr = %s and acct_session_id = %s'
                cur.execute(sql,(nas_addr,acct_session_id)) 
                online = cur.fetchone()
                if online:
                    ticket = _ticket(online) 
                    dsql = 'delete from slc_rad_online where nas_addr = %s and acct_session_id = %s'
                    cur.execute(dsql,(nas_addr,acct_session_id))
                    cur.execute(bsql,ticket)
                    conn.commit()  

        def _unlock_many():
            tickets = []
            with Connect(self.dbpool) as conn:
                cur = conn.cursor(cursors.DictCursor)
                cur.execute('select * from slc_rad_online where nas_addr = %s',(nas_addr,)) 
                for online in cur:
                    tickets.append(_ticket(online))
                if tickets:     
                    cur.executemany(bsql,tickets)
                    cur.execute('delete from slc_rad_online where nas_addr = %s',(nas_addr,))
                    conn.commit()                  

        if acct_session_id:_unlock_one()
        else:_unlock_many()


store = Store(DbPool())






