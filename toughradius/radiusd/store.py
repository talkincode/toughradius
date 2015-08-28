#!/usr/bin/env python
#coding=utf-8
from beaker.cache import CacheManager
from sqlalchemy.sql import text as _sql
import functools
import settings
import datetime

__cache_timeout__ = 600

cache = CacheManager(cache_regions={
      'short_term':{ 'type': 'memory', 'expire': __cache_timeout__ }
      }) 

###############################################################################
# Basic Define                                                            ####
###############################################################################

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


###############################################################################
# Database Store                                                          ####
###############################################################################        

class Store():

    def __init__(self,config,db_engine):
        global __cache_timeout__
        __cache_timeout__ = config.get("radiusd",'cache_timeout')
        self.db_engine = db_engine

    @cache.cache('get_param',expire=__cache_timeout__)   
    def get_param(self,param_name):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select param_value from  slc_param where param_name = :param_name"),param_name=param_name)
            param = cur.fetchone()
            return param and param['param_value'] or None

    def update_param_cache(self):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select param_name from  slc_param "))
            for param in cur:
                cache.invalidate(self.get_param,'get_param', str(param['param_name']))
                cache.invalidate(self.get_param,'get_param', unicode(param['param_name']))

    ###############################################################################
    # bas method                                                              ####
    ############################################################################### 

    def list_bas(self):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select * from  slc_rad_bas"))
            return [bas for bas in cur] 

    @cache.cache('get_bas',expire=__cache_timeout__)   
    def get_bas(self,ipaddr):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select * from slc_rad_bas where ip_addr = :ip_addr"),ip_addr=ipaddr)
            bas = cur.fetchone()
            return bas

    def update_bas_cache(self,ipaddr):
        cache.invalidate(self.get_bas,'get_bas',str(ipaddr))
        cache.invalidate(self.get_bas,'get_bas',unicode(ipaddr))

    ###############################################################################
    # user method                                                              ####
    ###############################################################################  

    @cache.cache('get_user',expire=__cache_timeout__)   
    def get_user(self,username):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select a.*,p.product_policy from slc_rad_account a,slc_rad_product p \
                where a.product_id = p.id and a.account_number = :account "),account=username)
            return  cur.fetchone()


    @cache.cache('get_user_attrs',expire=__cache_timeout__)   
    def get_user_attrs(self,username):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select * from slc_rad_account_attr where account_number = :account "),account=username)
            return cur.fetchall()


    @cache.cache('get_user_attr', expire=__cache_timeout__)
    def get_user_attr(self, username,attr_name):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("""select attr_value from slc_rad_account_attr
                                    where account_number = :account
                                    and attr_name = :attr_name"""),
                               account=username,attr_name=attr_name)
            b = cur.fetchone()
            return b and b['attr_value'] or None

    def get_user_balance(self,username):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select balance from slc_rad_account where account_number = :account "),account=username)
            b = cur.fetchone()  
            return b and b['balance'] or 0    
    
    def get_user_time_length(self,username):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select time_length from slc_rad_account where account_number = :account "),account=username)
            b = cur.fetchone()  
            return b and b['time_length'] or 0
    
    def get_user_flow_length(self,username):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select flow_length from slc_rad_account where account_number = :account "),account=username)
            b = cur.fetchone()  
            return b and b['flow_length'] or 0

    def update_user_cache(self,username):
        cache.invalidate(self.get_user,'get_user', str(username))
        cache.invalidate(self.get_user,'get_user', unicode(username))
        cache.invalidate(self.get_user_attrs,'get_user_attrs', str(username))
        cache.invalidate(self.get_user_attrs,'get_user_attrs', unicode(username))

    def update_user_balance(self,username,balance):
        with self.db_engine.begin() as conn:
            sql = _sql("update slc_rad_account set balance = :balance where account_number = :account")
            conn.execute(sql,balance=balance,account=username)
            self.update_user_cache(username)
            
    def update_user_time_length(self,username,time_length):
        with self.db_engine.begin() as conn:
            sql = _sql("update slc_rad_account set time_length = :time_length where account_number = :account")
            conn.execute(sql,time_length=time_length,account=username)
            self.update_user_cache(username)
    
    def update_user_flow_length(self,username,flow_length):
        with self.db_engine.begin() as conn:
            sql = _sql("update slc_rad_account set flow_length = flow_length where account_number = :account")
            conn.execute(sql,flow_length=flow_length,account=username)
            self.update_user_cache(username)

    def update_user_mac(self,username,mac_addr):
        with self.db_engine.begin() as conn:
            sql =_sql("update slc_rad_account set mac_addr = :mac_addr where account_number = :account")
            conn.execute(sql,mac_addr=mac_addr,account=username)
            self.update_user_cache(username)

    def update_user_vlan_id(self,username,vlan_id):
        with self.db_engine.begin() as conn:
            sql = _sql("update slc_rad_account set vlan_id = :vlan_id where account_number = :account")
            conn.execute(sql,vlan_id=vlan_id,account=username)
            self.update_user_cache(username)

    def update_user_vlan_id2(self,username,vlan_id2):
        with self.db_engine.begin() as conn:
            sql = _sql("update slc_rad_account set vlan_id2 = :vlan_id2 where account_number = :account")
            conn.execute(sql,vlan_id2=vlan_id2,account=username)   
            self.update_user_cache(username)     

    ###############################################################################
    # roster method                                                              ####
    ############################################################################### 

    @cache.cache('get_roster',expire=__cache_timeout__)   
    def get_roster(self,mac_addr):
        if mac_addr:
            mac_addr = mac_addr.upper()
        roster = None
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select * from slc_rad_roster where mac_addr = :mac_addr "),mac_addr=mac_addr)
            roster =  cur.fetchone()
        print roster
        if  roster:
            now = create_time = datetime.datetime.now()
            roster_start = datetime.datetime.strptime(roster['begin_time'],"%Y-%m-%d")
            roster_end = datetime.datetime.strptime(roster['end_time'],"%Y-%m-%d")
            if now < roster_start or now > roster_end:
                return None
            return roster
            
    def is_black_roster(self,mac_addr):
        roster = self.get_roster(mac_addr)
        return roster and roster['roster_type'] == 1 or False
        
    def is_white_roster(self,mac_addr):
        roster = self.get_roster(mac_addr)
        return roster and roster['roster_type'] == 0 or False
    
    def update_roster_cache(self,mac_addr):
        cache.invalidate(self.get_roster,'get_roster', str(mac_addr))
        cache.invalidate(self.get_roster,'get_roster', unicode(mac_addr))

    ###############################################################################
    # product method                                                         ####
    ############################################################################### 

    @cache.cache('get_product',expire=__cache_timeout__)   
    def get_product(self,product_id):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select * from slc_rad_product where id = :id "),id=product_id)
            return cur.fetchone()  

    @cache.cache('get_product_attrs',expire=__cache_timeout__)  
    def get_product_attrs(self,product_id):
        with self.db_engine.begin() as conn:
            cur = conn.execute(_sql("select * from slc_rad_product_attr where product_id = :product_id "),product_id=product_id)
            return cur.fetchall()  

    def update_product_cache(self,product_id):
        cache.invalidate(self.get_product,'get_product',product_id)
        cache.invalidate(self.get_product_attrs,'get_product_attrs',product_id)
        
    ###############################################################################
    # cache method                                                            ####
    ###############################################################################
    
    def update_all_cache(self):
        from beaker.cache import cache_managers
        for _cache in cache_managers.values():
            _cache.clear()


    ###############################################################################
    # online method                                                            ####
    ############################################################################### 
          
    def is_online(self,nas_addr,acct_session_id):
        with self.db_engine.begin() as conn:
            sql = _sql('select count(id) as online from slc_rad_online where  nas_addr = :nas_addr and acct_session_id = :session_id')
            cur = conn.execute(sql,nas_addr=nas_addr,session_id=acct_session_id)
            return cur.fetchone()['online'] > 0

    def count_online(self,account_number):
        with self.db_engine.begin() as conn:
            sql = _sql('select count(id) as online from slc_rad_online where  account_number = :account')
            cur = conn.execute(sql,account=account_number) 
            return cur.fetchone()['online']

    def get_online(self,nas_addr,acct_session_id):
        with self.db_engine.begin() as conn:
            sql = _sql('select * from slc_rad_online where  nas_addr = :nas_addr and acct_session_id = :session_id')
            cur = conn.execute(sql,nas_addr=nas_addr,session_id=acct_session_id)
            return cur.fetchone()     

    def get_nas_onlines(self,nas_addr):
        with self.db_engine.begin() as conn:
            sql = _sql('select * from slc_rad_online where nas_addr = :nas_addr')
            cur = conn.execute(sql,nas_addr=nas_addr) 
            return cur.fetchall()

    def get_nas_onlines_byuser(self, account_number):
        with self.db_engine.begin() as conn:
            sql = _sql('select * from slc_rad_online where  account_number = :account')
            cur = conn.execute(sql, account=account_number)
        return cur.fetchall()

    def add_online(self,online):
        with self.db_engine.begin() as conn:
            keys = ','.join(online.keys())
            vals = ",".join(["'%s'"% c for c in online.values()])
            sql = _sql('insert into slc_rad_online (%s) values(%s)'%(keys,vals))
            conn.execute(sql)
            
    
    def check_online_over(self):
        onlines = []
        with self.db_engine.begin() as conn:
            sql = _sql('select acct_start_time,nas_addr,acct_session_id from slc_rad_online')
            onlines = conn.execute(sql)
        
        for online in onlines:
            start_time = datetime.datetime.strptime(online['acct_start_time'],"%Y-%m-%d %H:%M:%S")
            _datetime = datetime.datetime.now() 
            if (_datetime - start_time).seconds > 3600 * 3:
                self.unlock_online(
                    online['nas_addr'],
                    online['acct_session_id'],
                    settings.STATUS_TYPE_CHECK_ONLINE
                )
            
              
            
    def update_online(self,online):
        with self.db_engine.begin() as conn:
            online_sql = _sql("""update slc_rad_online set 
                billing_times = :billing_times,
                input_total = :input_total,
                output_total = :output_total
                where nas_addr = :nas_addr and acct_session_id = :session_id
            """)
            conn.execute(online_sql,
                billing_times = online['billing_times'],
                input_total = online['input_total'],
                output_total = online['output_total'],
                nas_addr = online['nas_addr'],
                session_id = online['acct_session_id']
            )
            

    def update_billing(self,billing,time_length=0,flow_length=0):
        with self.db_engine.begin() as conn:
            # update account
            balan_sql = _sql("""update slc_rad_account set 
                balance = :balance,
                time_length=:time_length,
                flow_length=:flow_length
                where account_number = :account
            """)
            conn.execute(balan_sql,
                balance=billing.balance,
                time_length=time_length,
                flow_length=flow_length,
                account=billing.account_number
            )
            
            # update online
            online_sql = _sql("""update slc_rad_online set 
                billing_times = :billing_times,
                input_total = :input_total,
                output_total = :output_total
                where nas_addr = :nas_addr and acct_session_id = :session_id
            """)
            conn.execute(online_sql,
                billing_times=billing.acct_session_time,
                input_total=billing.input_total,
                output_total=billing.output_total,
                nas_addr=billing.nas_addr,
                session_id=billing.acct_session_id
            )
            
            # update billing
            keys = ','.join(billing.keys())
            vals = ",".join(["'%s'"% c for c in billing.values()])
            billing_sql = _sql('insert into slc_rad_billing (%s) values(%s)'%(keys,vals))
            conn.execute(billing_sql)
            
            
        self.update_user_cache(billing.account_number) 
    
    def del_online(self,nas_addr,acct_session_id):
        with self.db_engine.begin() as conn:
            sql = _sql('delete from slc_rad_online where nas_addr = :nas_addr and acct_session_id = :session_id')
            conn.execute(sql,nas_addr=nas_addr,session_id=acct_session_id)
            

    def add_ticket(self,ticket):
        _ticket = ticket.copy()
        for _key in _ticket:
            if _key not in ticket_fds:
                del ticket[_key]
        with self.db_engine.begin() as conn:
            keys = ','.join(ticket.keys())
            vals = ",".join(["'%s'"% c for c in ticket.values()])
            sql = _sql('insert into slc_rad_ticket (%s) values(%s)'%(keys,vals))
            conn.execute(sql)
            

    def unlock_online(self,nas_addr,acct_session_id,stop_source):
        bsql = _sql(''' insert into slc_rad_ticket 
                    (
                        account_number,
                        acct_session_id,
                        acct_start_time,
                        nas_addr,
                        framed_ipaddr,
                        start_source,
                        acct_session_time,
                        acct_stop_time,
                        stop_source
                    ) values(
                        :account_number,
                        :acct_session_id,
                        :acct_start_time,
                        :nas_addr,
                        :framed_ipaddr,
                        :start_source,
                        :acct_session_time,
                        :acct_stop_time,
                        :stop_source
                    )''') 
                    
        def _ticket(online):
            _datetime = datetime.datetime.now()
            _starttime = datetime.datetime.strptime(online['acct_start_time'],"%Y-%m-%d %H:%M:%S")
            session_time = (_datetime - _starttime).seconds
            stop_time = _datetime.strftime( "%Y-%m-%d %H:%M:%S")
            return dict(
                account_number = online['account_number'],
                acct_session_id = online['acct_session_id'],
                acct_start_time = online['acct_start_time'],
                nas_addr = online['nas_addr'],
                framed_ipaddr = online['framed_ipaddr'],
                start_source = online['start_source'], 
                acct_session_time = session_time,
                acct_stop_time = stop_time,
                stop_source = stop_source
            )

        def _unlock_one():
            ticket = None
            with self.db_engine.begin() as conn:
                sql = _sql('select * from slc_rad_online where  nas_addr = :nas_addr and acct_session_id = :session_id')
                cur = conn.execute(sql,nas_addr=nas_addr,session_id=acct_session_id)
                online = cur.fetchone()
                if online:
                    ticket = _ticket(online) 
                    dsql = _sql('delete from slc_rad_online where nas_addr = :nas_addr and acct_session_id = :session_id')
                    conn.execute(dsql,nas_addr=nas_addr,session_id=acct_session_id)
                    conn.execute(bsql,**ticket)
                      

        def _unlock_many():
            tickets = []
            with self.db_engine.begin() as conn:
                cur = conn.execute(_sql('select * from slc_rad_online where nas_addr = :nas_addr'),nas_addr=nas_addr) 
                for online in cur:  
                    conn.execute(bsql,**_ticket(online))
                    conn.execute(_sql('delete from slc_rad_online where nas_addr = :nas_addr'),nas_addr=nas_addr)
                                      

        if acct_session_id:_unlock_one()
        else:_unlock_many()

    

