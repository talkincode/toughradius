#!/usr/bin/env python
#coding:utf-8
import redis
from gevent import socket
redis.connection.socket = socket
from .base import BasicAdapter
from toughradius.common import rediskeys
from toughradius.common.rediskeys import UserAttrs, NasAttrs
from toughradius.common.rediskeys import UserStates, NasStates
from toughradius.txradius import message
import datetime
import logging
import time


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

    def getClients(self):
        naslist = self.redis.zrange(rediskeys.NasSetKey,0,-1) or set()
        clients = {}
        for naskey in naslist:
            nas = self.redis.hgetall(naskey)
            ipaddr = nas.get(NasAttrs.ipaddr.name)
            nasid = nas.get(NasAttrs.nasid.name)
            if nasid:
                clients[nasid] = nas
            if ipaddr:
                clients[ipaddr] = nas
        return clients

    def processAuth(self,req):
        # check exists
        username = req.get_user_name()
        user = self.redis.hgetall(rediskeys.UserHKey(username))
        if not user and self.config.radiusd.bypass_userpwd == 1:
            raise RedisAdapterError('user {0} not exists'.format(username))

        #check pause
        if user.get(UserAttrs.status.name) == 0:
            raise RedisAdapterError('user is pause')

        # check  expire
        expire_date = user.get(UserAttrs.expire_date.name)
        expire_time = user.get(UserAttrs.expire_time.name)
        if self.is_expire(expire_date,expire_time):
            raise RedisAdapterError('user is expire at {0} {1}'.format(expire_date,expire_time))

        # check password
        password = user.get(UserAttrs.password.name)
        if self.config.radiusd.bypass_pwd == 0 and  self.config.radiusd.bypass_userpwd == 0:
            if user.get(UserAttrs.bypass_pwd.name, 1) == 1:
                if not req.is_valid_pwd(password):
                    raise RedisAdapterError('user password error')

        # check mac bind
        req_mac_addr = req.get_mac_addr()
        if not self.check_mac_bind(user,req_mac_addr):
            raise RedisAdapterError('user mac bind error req={0}, bind={1}'.format(
                req_mac_addr,user.get(UserAttrs.mac_addr.name)))

        pre_reply = dict(code=0,msg='ok')
        pre_reply['input_rate'] = user.get(UserAttrs.input_rate.name,0)
        pre_reply['output_rate'] = user.get(UserAttrs.output_rate.name,0)
        pre_reply.setdefault('attrs',{})['Session-Timeout'] = self.calc_session_time(
            user.get(UserAttrs.bill_type.name,'day'),
            expire_date,
            expire_time,
            user.get(UserAttrs.time_amount.name,0)
        )

        user_attrs = self.redis.hgetall(rediskeys.UserRadAttrsHKey(username)) or {}
        for _name, _value in user_attrs.iteritems():
            pre_reply['attrs'][_name] = _value

        return pre_reply

    def processAcct(self, req):
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
        nasaddr = req.get_nas_addr()
        sessionid = req.get_acct_sessionid()
        online_key = rediskeys.OnlineHKey(nasid,nasaddr,sessionid)
        if self.redis.exists(online_key):
            raise RedisAdapterError('user {0} session duplicate'.format(username))

        # check online limit 
        online_limit = int(self.redis.hget(online_key,UserAttrs.online_limit.name) or 0)
        if online_limit > 0:
            online_count = int(self.redis.scard(rediskeys.UserOnlineSetKey(username)) or 0)
            if online_count > online_limit:
                raise RedisAdapterError('user {0} online limit {1}'.format(username,online_limit))

        billing = req.get_billing()
        billing['pub_nas_addr'] = req.source[0]
        score =int(time.time())
        with self.redis.pipeline() as pipe:
            pipe.hmset(online_key,billing)
            pipe.zadd(rediskeys.OnlineSetKey,score,online_key)
            pipe.zadd(rediskeys.UserOnlineSetKey(username),score,online_key)
            pipe.zadd(rediskeys.NasOnlineSetKey(nasid,sessionid),score,online_key)
            pipe.execute()
        logging.info(u'user {0} start billing'.format(username))
        return dict(code=0, msg='ok')


    def acct_update(self,req):
        username = req.get_user_name()
        if not self.redis.exists(rediskeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))

        nasid = req.get_nas_id()
        nasaddr = req.get_nas_addr()
        sessionid = req.get_acct_sessionid()
        online_key = rediskeys.OnlineHKey(nasid,nasaddr,sessionid)
        if not self.redis.exists(online_key):
            billing = req.get_billing()
            billing['pub_nas_addr'] = req.source[0]
            score = int(time.time())
            with self.redis.pipeline() as pipe:
                pipe.hmset(online_key,billing)
                pipe.zadd(rediskeys.OnlineSetKey,score, online_key)
                pipe.zadd(rediskeys.UserOnlineSetKey(username),score, online_key)
                pipe.zadd(rediskeys.NasOnlineSetKey(nasid,nasaddr),score, online_key)
                pipe.execute()
            logging.info(u'add user {0} billing data on update'.format(username))
        else:
            self.billing(req)
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

        nasid = req.get_nas_id()
        nasaddr = req.get_nas_addr()
        sessionid = req.get_acct_sessionid()
        online_key = rediskeys.OnlineHKey(nasid,nasaddr,sessionid)
        self.billing(req)
        with self.redis.pipeline() as pipe:
            pipe.delete(online_key)
            pipe.zrem(rediskeys.OnlineSetKey,online_key)
            pipe.zrem(rediskeys.UserOnlineSetKey(username),online_key)
            pipe.zrem(rediskeys.NasOnlineSetKey(nasid,nasaddr),online_key)
            pipe.execute()
        logging.info(u'delete online user {0}'.format(username))
        return dict(code=0, msg='ok')

    def billing(self,req):
        username = req.get_user_name()
        try:
            user_key = rediskeys.UserHKey(username)
            nasid = req.get_nas_id()
            nasaddr = req.get_nas_addr()
            sessionid = req.get_acct_sessionid()
            online_key = rediskeys.OnlineHKey(nasid, nasaddr, sessionid)
            if self.redis.hget(user_key,UserAttrs.bill_type.name) not in ('second',):
                return

            if not self.redis.exists(online_key):
                return

            oldsession_time = self.redis.hget(online_key,'acct_session_time')
            if not oldsession_time:
                return

            use_time = req.get_acct_sessiontime() - oldsession_time
            assert use_time > 0
            if (self.redis.hget(user_key,UserAttrs.time_amount.name) or 0) < 0:
                self.redis.hset(user_key,UserAttrs.time_amount.name,0)
            else:
                self.redis.hincrby(user_key,UserAttrs.time_amount.name,-use_time)
        except:
            logging.error('billing error for user {}'.format(username),exc_info=True)


    def check_mac_bind(self, user, req_mac_addr):
        if int(user.get(UserAttrs.bind_mac.name,0)) == 1:
            mac_addr = user.get(UserAttrs.mac_addr.name)
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
        session_timeout = 0
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











