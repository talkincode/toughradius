#!/usr/bin/env python
#coding=utf-8
import yaml
import MySQLdb
from MySQLdb import cursors
from DBUtils.PooledDB import PooledDB

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

dbpool = DbPool()

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
                "where a.product_id = p.product_id and a.account_number = %s ",(username,))
            user =  cur.fetchone()
            return user

    def update_user_mac(self,username,mac_addr):
        with Cursor(self.dbpool) as cur:
            sql = "update slc_rad_account set mac_addr = %s where account_number = %s"
            cur.execute(sql,(mac_addr,username))

    def update_user_vlan_id(self,username,vlan_id):
        with Cursor(self.dbpool) as cur:
            sql = "update slc_rad_account set vlan_id = %s where account_number = %s"
            cur.execute(sql,(vlan_id,username))

    def update_user_vlan_id2(self,username,vlan_id2):
        with Cursor(self.dbpool) as cur:
            sql = "update slc_rad_account set vlan_id2 = %s where account_number = %s"
            cur.execute(sql,(vlan_id2,username))            


    def get_group(self,group_id):
        with Cursor(self.dbpool) as cur:
            cur.execute("select * from slc_rad_group where id = %s ",(group_id,))
            grp =  cur.fetchone()
            return grp


store = Store(dbpool)




