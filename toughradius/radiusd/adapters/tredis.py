#!/usr/bin/env python
#coding:utf-8
from gevent import socket
from gevent import spawn_later
import redis
redis.connection.socket = socket
from .base import BasicAdapter
from toughradius.common import tools
from toughradius.common import rediskeys
from toughradius.txradius import message
from hashlib import md5
import datetime
import logging



class RedisAdapterError(BaseException):pass

class RedisAdapter(BasicAdapter):

    def __init__(self, config):
        BasicAdapter.__init__(self,config)
        self.redisconfig = self.config.adapters.redis
        _pool = redis.ConnectionPool(
            host=self.redisconfig.get('host'),
            port=self.redisconfig.get("port"),
            password=self.redisconfig.get('passwd'),db=0,
            max_connections=int(self.redisconfig.get('pool_size',100))
        )
        self.redis = redis.StrictRedis(connection_pool=_pool)

    def get_clients(self):
        naslist = self.redis.smembers(rediskeys.NasSetKey) or set()
        clients = {}
        for nas in naslist:
            ipaddr = nas.get('ipaddr')
            nasid = nas.get('nasid')
            if nasid:
                clients[nasid] = nas
            if ipaddr:
                clients[ipaddr] = nas
        return clients

    def auth(self,req):
        # check exists
        username = req.get_user_name()
        user = self.redis.hgetall(rediskeys.UserHKey(username))
        if not user and self.config.radiusd.bypass_userpwd == 1:
            raise RedisAdapterError('user {0} not exists'.format(username))

        #check pause
        if user.get('status') == 0:
            raise RedisAdapterError('user is pause')

        # check  expire
        expire_date = user.get('expire_date')
        expire_time = user.get('expire_time')
        if self.is_expire(expire_date,expire_time):
            raise RedisAdapterError('user is expire at {0} {1}'.format(expire_date,expire_time))

        # check password
        password = user.get('password')
        if self.config.radiusd.bypass_pwd == 0 and  self.config.radiusd.bypass_userpwd == 0:
            if user.get('bypass_pwd', 1) == 1:
                if not req.is_valid_pwd(password):
                    raise RedisAdapterError('user password error')

        # check mac bind
        req_mac_addr = self.get_mac_addr()
        if not self.check_mac_bind(user,req_mac_addr):
            raise RedisAdapterError('user mac bind error req={0}, bind={1}'.format(
                req_mac_addr,user.get('mac_addr')))

        pre_reply = dict(code=0,msg='ok')
        pre_reply['input_rate'] = user.get('input_rate',0)
        pre_reply['output_rate'] = user.get('output_rate',0)
        pre_reply['attrs']['Session-Timeout'] = self.calc_session_time(
            user.get('bill_type','day'),
            expire_date,
            expire_time,
            user.get('time_amount',0)
        )

        for attrname, attrvalue in (self.redis.hgetall(rediskeys.UserRadAttrsHKey(username)) or {}):
            pre_reply['attrs'][attrname] = attrvalue

        return pre_reply

    def acct(self,req):
        status_type = req.get_acct_status_type()
        if status_type == message.STATUS_TYPE_START:
            return self.acct_start(req)
        elif status_type == message.STATUS_TYPE_UPDATE:
            return self.acct_update(req)
        elif status_type == message.STATUS_TYPE_STOP:
            return self.acct_stop(req)
        
    def acct_start(self,req):
        username = req.get_user_name()
        if not self.redis.exists(rediskeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))

        nasid = req.get_nas_id()
        sessionid = req.get_acct_sessionid()
        online_key = rediskeys.OnlineHKey(nasid,sessionid)
        if self.redis.exists(online_key):
            raise RedisAdapterError('user {0} is online'.format(username))

        billing = req.get_billing()
        billing['pub_nas_addr'] = req.source[0]
        self.redis.hmset(online_key,billing)
        logging.info(u'user {0} start billing'.format(username))
        return dict(code=0, msg='ok')


    def acct_update(self,req):
        username = req.get_user_name()
        if not self.redis.exists(rediskeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))

        online_key = rediskeys.OnlineHKey(req.get_nas_id(), req.get_acct_sessionid())
        if not self.redis.exists(online_key):
            billing = req.get_billing()
            billing['pub_nas_addr'] = req.source[0]
            self.redis.hmset(online_key,billing)
            logging.info(u'add user {0} billing data on update'.format(username))
        else:
            self.billing(req, online_key)
            self.redis.hmset(online_key,dict(
                acct_session_time=req.get_acct_sessiontime(),
                acct_input_total=req.get_input_total(),
                acct_output_total=req.get_output_total(),
                acct_input_packets=req.get_acct_input_packets(),
                acct_output_packets=req.get_acct_output_packets()
            ))

        return dict(code=0, msg='ok')


    def acct_stop(self,req):
        username = req.get_user_name()
        if not self.redis.exists(rediskeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))
        online_key = rediskeys.OnlineHKey(req.get_nas_id(), req.get_acct_sessionid())
        self.billing(req)
        self.redis.hdel(online_key)
        logging.info(u'delete online user {0}'.format(username))
        return dict(code=0, msg='ok')

    def billing(self,req):
        username = req.get_user_name()
        try:
            user_key = rediskeys.UserHKey(username)
            online_key = rediskeys.OnlineHKey(req.get_nas_id(), req.get_acct_sessionid())
            if self.redis.hget(user_key,'bill_type') not in ('second',):
                return

            if not self.redis.exists(online_key):
                return

            oldsession_time = self.redis.hget(online_key,'acct_session_time')
            if not oldsession_time:
                return

            use_time = req.get_acct_sessiontime() - oldsession_time
            assert use_time > 0
            if (self.redis.hget(user_key,'time_amount') or 0) < 0:
                self.redis.hset(user_key,'time_amount',0)
            else:
                self.redis.hincrby(user_key,'time_amount',-use_time)
        except:
            logging.error('billing error for user {}'.format(username),exc_info=True)


    def check_mac_bind(self, user, req_mac_addr):
        if int(user.get('bind_mac',0)) == 1:
            mac_addr = user.get('mac_addr')
            if mac_addr and mac_addr != req_mac_addr:
                return False
        return True

    def is_expire(self,expire_date,expire_time):
        if not expire_time:
            expire_time = '23:59:59'
        if expire_date:
            try:
                _timestr = "%s %s" % (expire_date,expire_time)
                _expire = datetime.datetime.strptime(_timestr, "%Y-%m-%d %H:%M:%S")
                return _expire < datetime.datetime.now()
            except:
                import traceback
                traceback.print_exc()
                return False
        return False


    def calc_session_time(self,bill_type, expire_date,expire_time, time_amount):
        if bill_type == 'day':
            if not expire_time:
                expire_time = '23:59:59'
            datetimestr = "%s %s" % (expire_date,expire_time)
            now_datetime = datetime.datetime.now()
            expire_datetime = datetime.datetime.strptime(datetimestr,"%Y-%m-%d %H:%M:%S")
            session_timeout = (expire_datetime - now_datetime).total_seconds() 
        elif bill_type == 'second':
            session_timeout = time_amount

        if session_timeout < 0:
            session_timeout = 0

        return session_timeout











