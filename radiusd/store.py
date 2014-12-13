#!/usr/bin/env python
#coding=utf-8
import yaml
import MySQLdb
from MySQLdb import cursors
from DBUtils.PooledDB import PooledDB

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
    def __init__(self,config="config.yaml"):
        with open(config,'rb') as cf:
            self.dbpool = PooledDB(creator=MySQLdb,**yaml.load(cf)['mysql'])

    def connect(self):
        return self.dbpool.connection()

class Store():

    def __init__(self,dbpool=None):
        self.dbpool = dbpool 

    def list_bas(self):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from  slc_rad_bas")
            return [bas for bas in cur] 

    def get_user(self,username):
        with Cursor(self.dbpool) as cur:
            cur.execute("select a.*,p.product_policy from slc_rad_account a,slc_rad_product p "
                "where a.product_id = p.id and a.account_number = %s ",(username,))
            user =  cur.fetchone()
            return user

    def update_user_balance(self,username,balance):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set balance = %s where account_number = %s"
            cur.execute(sql,(balance,username))
            conn.commit()

    def update_user_mac(self,username,mac_addr):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set mac_addr = %s where account_number = %s"
            cur.execute(sql,(mac_addr,username))
            conn.commit()

    def update_user_vlan_id(self,username,vlan_id):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set vlan_id = %s where account_number = %s"
            cur.execute(sql,(vlan_id,username))
            conn.commit()

    def update_user_vlan_id2(self,username,vlan_id2):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = "update slc_rad_account set vlan_id2 = %s where account_number = %s"
            cur.execute(sql,(vlan_id2,username))    
            conn.commit()        


    def get_group(self,group_id):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_group where id = %s ",(group_id,))
            return cur.fetchone()

    def get_product(self,product_id):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_product where id = %s ",(product_id,))
            return cur.fetchone()     
            
    def get_online(self,nas_addr,sessionid):
        with Cursor(self.dbpool) as cur: 
            sql = 'select * from slc_rad_online where  nas_addr = %s and sessionid = %s'
            cur.execute(sql,(nas_addr,sessionid)) 
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
    
    def del_online(self,nas_addr,sessionid):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = 'delete from slc_rad_online where nas_addr = %s and sessionid = %s'
            cur.execute(sql,(nas_addr,sessionid))
            conn.commit()

    def del_nas_onlines(self,nas_addr):
        with Connect(self.dbpool) as conn:
            cur = conn.cursor()
            sql = 'delete from slc_rad_online where nas_addr = %s'
            cur.execute(sql,(nas_addr,))
            conn.commit()
            
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
            
store = Store(DbPool())






