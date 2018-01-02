#!/usr/bin/env python
# coding:utf-8
from gevent import socket

try:
    import redis
    redis.connection.socket = socket
except:
    pass
from .base import BasicAdapter
from toughradius.pyrad import message
import datetime
import logging
import time


class RedisKeys(object):

    NasIdHKey = "toughradius:nasid_hkey:{nasid}".format
    NasIpHKey = "toughradius:nasip_hkey:{nasip}".format
    NasSetKey = "toughradius:nas_set_key"

    UserSetKey = "toughradius:user_set_key"
    UserHKey = "toughradius:user_hkey:{username}".format
    UserExtAttrsHKey = "toughradius:user_ext_attrs_hkey:{username}".format
    UserRadAttrsHKey = "toughradius:user_radius_attrs_hkey:{username}".format

    OnlineHKey = "toughradius:online_hkey:{nasid}:{username}:{sessionid}".format
    OnlineSetKey = "toughradius:online_set_key"
    UserOnlineSetKey = "toughradius:user_online_set_key:{username}".format
    NasOnlineSetKey = "toughradius:nas_online_set_key:{nasid}".format

class RadiusUtils(object):

    @staticmethod
    def check_mac_bind(user, req_mac_addr):
        if int(user.get('bind_mac', 0)) == 1:
            mac_addr = user.get('mac_addr')
            if mac_addr and mac_addr != req_mac_addr:
                return False
        return True

    @staticmethod
    def is_expire(expire_time):
        if expire_time:
            try:
                _expire = datetime.datetime.strptime(expire_time, "%Y-%m-%d %H:%M:%S")
                return _expire < datetime.datetime.now()
            except:
                import traceback
                traceback.print_exc()
                return False
        return False

    @staticmethod
    def calc_session_time(expire_time):
        now_datetime = datetime.datetime.now()
        expire_datetime = datetime.datetime.strptime(expire_time, "%Y-%m-%d %H:%M:%S")
        session_timeout = (expire_datetime - now_datetime).total_seconds()
        if session_timeout < 0:
            session_timeout = 0
        return session_timeout


class RedisAdapterError(BaseException): pass


class RedisAdapter(BasicAdapter):
    def __init__(self, config):
        BasicAdapter.__init__(self, config)
        self.redisconfig = self.settings.ADAPTERS['redis']
        _pool = redis.ConnectionPool(
            host=self.redisconfig.get('host'),
            port=self.redisconfig.get("port"),
            password=self.redisconfig.get('passwd'),
            db=self.redisconfig.get('db',0),
            max_connections=int(self.redisconfig.get('pool_size', 100))
        )
        self.redis = redis.StrictRedis(connection_pool=_pool)

    def getClient(self, nasip=None, nasid=None):
        if nasip:
            nas = self.redis.get(RedisKeys.NasIpHKey(nasip=nasip))
            if nas:
                return nas
        elif nasid:
            nas = self.redis.get(RedisKeys.NasIdHKey(nasid=nasid))
            if nas:
                return nas

    def processAuth(self, req):
        # check exists
        username = req.get_user_name()
        user = self.redis.hgetall(RedisKeys.UserHKey(username))
        if not user:
            raise RedisAdapterError('user {0} not exists'.format(username))

        # check pause
        if user.get('status') in ('0', 'disabled'):
            raise RedisAdapterError('user is disabled')

        # check  expire
        expire_time = user.get('expire_time')
        if RadiusUtils.is_expire(expire_time):
            raise RedisAdapterError('user is expire at {0}'.format(expire_time))

        # check password
        password = user.get('password')
        if int(self.settings.RADIUSD.get('ignore_passwd', 1)) == 1:
            if user.get('ignore_passwd', 1) == 1:
                if not req.is_valid_pwd(password):
                    raise RedisAdapterError('user password error')

        # check mac bind
        req_mac_addr = req.get_mac_addr()
        if not RadiusUtils.check_mac_bind(user, req_mac_addr):
            raise RedisAdapterError('user mac bind error req={0}, bind={1}'.format(
                req_mac_addr, user.get('mac_addr')))

        pre_reply = dict(code=0, msg='ok')
        pre_reply['ext_attrs']['input_rate'] = user.get('input_rate', 0)
        pre_reply['ext_attrs']['output_rate'] = user.get('output_rate', 0)
        pre_reply['radius_attrs']['Session-Timeout'] = RadiusUtils.calc_session_time(expire_time)

        for attrname, attrvalue in (self.redis.hgetall(RedisKeys.UserExtAttrsHKey(username)) or {}):
            pre_reply['ext_attrs'][attrname] = attrvalue

        for attrname, attrvalue in (self.redis.hgetall(RedisKeys.UserRadAttrsHKey(username)) or {}):
            pre_reply['radius_attrs'][attrname] = attrvalue

        return pre_reply

    def processAcct(self, req):
        status_type = req.get_acct_status_type()
        if status_type == message.STATUS_TYPE_START:
            return self.accounting_start(req)
        elif status_type == message.STATUS_TYPE_UPDATE:
            return self.accounting_update(req)
        elif status_type == message.STATUS_TYPE_STOP:
            return self.accounting_stop(req)
        elif status_type in (message.STATUS_TYPE_ACCT_ON, message.STATUS_TYPE_ACCT_OFF):
            return self.accounting_onoff(req)

    def accounting_start(self, req):
        username = req.get_user_name()
        if not self.redis.exists(RedisKeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))

        nasid = req.get_nas_id()
        sessionid = req.get_acct_sessionid()
        online_key = RedisKeys.OnlineHKey(nasid, username, sessionid)
        if self.redis.exists(online_key):
            raise RedisAdapterError('user {0} session duplicate'.format(username))

        # check online limit
        online_limit = int(self.redis.hget(online_key, "online_limit") or 0)
        if online_limit > 0:
            online_count = int(self.redis.scard(RedisKeys.UserOnlineSetKey(username)) or 0)
            if online_count > online_limit:
                raise RedisAdapterError('user {0} online limit {1}'.format(username, online_limit))

        billing = req.get_billing()
        billing['pub_nas_addr'] = req.source[0]
        score = int(time.time())
        with self.redis.pipeline() as pipe:
            pipe.hmset(online_key, billing)
            pipe.zadd(RedisKeys.OnlineSetKey, score, online_key)
            pipe.zadd(RedisKeys.UserOnlineSetKey(username), score, online_key)
            pipe.zadd(RedisKeys.NasOnlineSetKey(nasid), score, online_key)
            pipe.execute()
        logging.info(u'user {0} start billing'.format(username))
        return dict(code=0, msg='ok')

    def accounting_update(self, req):
        username = req.get_user_name()
        if not self.redis.exists(RedisKeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))

        nasid = req.get_nas_id()
        sessionid = req.get_acct_sessionid()
        online_key = RedisKeys.OnlineHKey(nasid, username, sessionid)
        if not self.redis.exists(online_key):
            billing = req.get_billing()
            billing['pub_nas_addr'] = req.source[0]
            score = int(time.time())
            with self.redis.pipeline() as pipe:
                pipe.hmset(online_key,billing)
                pipe.zadd(RedisKeys.OnlineSetKey,score, online_key)
                pipe.zadd(RedisKeys.UserOnlineSetKey(username),score, online_key)
                pipe.zadd(RedisKeys.NasOnlineSetKey(nasid),score, online_key)
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

    def accounting_stop(self, req):
        username = req.get_user_name()
        if not self.redis.exists(RedisKeys.UserHKey(username)):
            raise RedisAdapterError('user {0} not exists'.format(username))

        nasid = req.get_nas_id()
        sessionid = req.get_acct_sessionid()
        online_key = RedisKeys.OnlineHKey(nasid, username, sessionid)
        self.billing(req)
        with self.redis.pipeline() as pipe:
            pipe.delete(online_key)
            pipe.zrem(RedisKeys.OnlineSetKey, online_key)
            pipe.zrem(RedisKeys.UserOnlineSetKey(username), online_key)
            pipe.zrem(RedisKeys.NasOnlineSetKey(nasid), online_key)
            pipe.execute()
        logging.info(u'delete online user {0}'.format(username))
        return dict(code=0, msg='ok')

    def accounting_onoff(self, req):
        nasid = req.get_nas_id()
        try:
            online_keys = self.redis.smembers(RedisKeys.NasOnlineSetKey(nasid))
            delkeys = set()
            remkeys = set()
            for online_key in online_keys:
                username = online_key.split(':')[3]
                delkeys.add(online_key)
                remkeys.add(RedisKeys.OnlineSetKey)
                remkeys.add(RedisKeys.NasOnlineSetKey(nasid))
                remkeys.add(RedisKeys.UserOnlineSetKey(username))

            with self.redis.pipeline() as pipe:
                for delkey in delkeys:
                    pipe.delete(delkey)
                    for remkey in delkeys:
                        pipe.zrem(remkey, delkey)
                pipe.execute()
        except:
            logging.error('accounting on off error for {0}'.format(nasid), exc_info=True)

        return dict(code=0, msg='ok')

    def billing(self, req):
        username = req.get_user_name()
        try:
            user_key = RedisKeys.UserHKey(username)
            online_key = RedisKeys.OnlineHKey(req.get_nas_id(), req.get_acct_sessionid())
            if self.redis.hget(user_key, 'bill_type') not in ('time',):
                return

            if not self.redis.exists(online_key):
                return

            if self.redis.hget(user_key, 'bill_type') == 'flow':
                acct_input_total = self.redis.hget(online_key, 'acct_input_total')
                acct_output_total = self.redis.hget(online_key, 'acct_output_total')
                if acct_output_total is None:
                    return

                use_flows = req.get_acct_output_total() - int(acct_output_total)
                if use_flows < 0:
                    use_flows = 0
                if (self.redis.hget(user_key, 'flow_length') or 0) < 0:
                    self.redis.hset(user_key, 'flow_length', 0)
                    self.notify_user_flow(username, use_flows)
                else:
                    self.redis.hincrby(user_key, 'flow_length', -use_flows)
                    self.notify_user_flow(username, use_flows)
        except:
            logging.error('billing error for user {}'.format(username), exc_info=True)


    def notify_user_flow(self, username, useflows):
        pass


adapter = RedisAdapter





