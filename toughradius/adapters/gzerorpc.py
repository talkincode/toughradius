#!/usr/bin/env python
# coding:utf-8
try:
    import zerorpc
except:
    pass
from toughradius.common.mcache import cache
from .base import BasicAdapter
from toughradius.pyrad import message
import datetime
import logging
import time

class RadiusUtils(object):

    @staticmethod
    def check_mac_bind(user, req_mac_addr):
        if user.bind_mac == 1:
            if user.mac_address and user.mac_address != req_mac_addr:
                return False
        return True

    @staticmethod
    def check_in_vlan_bind(user, req_vlan):
        if user.bind_in_vlan == 1:
            if user.in_vlan and user.in_vlain != req_vlan:
                return False
        return True

    @staticmethod
    def check_out_vlan_bind(user, req_vlan):
        if user.bind_out_vlan == 1:
            if user.out_vlan and user.out_vlain != req_vlan:
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

class RadiusUser(object):

    def __init__(self,user):
        self.src = user
        self.username = user.get('username')
        self.password = user.get('password')
        self.bill_type = user.get('bill_type')
        self.expire_time = user.get('expire_time')
        self.ignore_password = int(user.get('ignore_password',0))
        self.online_limit = int(user.get('online_limit',0))
        self.mac_address = user['attrs'].get('MAC_ADDRESS')
        self.in_vlan = int(user['attrs'].get('IN_VLAN',0))
        self.out_vlan = int(user['attrs'].get('OUT_VLAN',0))
        self.bind_mac = int(user['attrs'].get('BIND_MAC',0))
        self.bind_in_vlan = int(user['attrs'].get('BIND_IN_VLAN',0))
        self.bind_out_vlan = int(user['attrs'].get('BIND_OUT_VLAN',0))
        self.input_rate = int(user['attrs'].get('INPUT_RATE',0))
        self.output_rate = int(user['attrs'].get('OUTPUT_RATE',0))
        self.radius_attrs = user.get('radius_attrs',{})


class RedisAdapter(BasicAdapter):
    def __init__(self, config):
        BasicAdapter.__init__(self, config)
        self.zrpc = zerorpc.Client(connect_to=self.settings.ADAPTERS['zerorpc']['connect'])

    def getClient(self, nasip=None, nasid=None):
        def fetch_result():
            return self.zrpc.radius_find_nas(nasip,nasid)
        return cache.aget('toughradius.nas.cache.{0}.{1}'.format(nasid, nasip), fetch_result, expire=30)

    def processAuth(self, req):
        # check exists
        username = req.get_user_name()
        _user = self.zrpc.radius_find_user(username, async=True).get()
        if not _user:
            raise RedisAdapterError('user {0} not exists or user status not enabled'.format(username))

        user = RadiusUser(_user)

        # check  expire
        if RadiusUtils.is_expire(user.expire_time):
            raise RedisAdapterError('user is expire at {0}'.format(user.expire_time))

        # check password
        if int(self.settings.RADIUSD.get('ignore_password', 0)) == 1:
            if user.ignore_password == 1:
                if not req.is_valid_pwd(user.password):
                    raise RedisAdapterError('user password error')

        # check mac bind
        req_mac_addr = req.get_mac_addr()
        if not RadiusUtils.check_mac_bind(user, req_mac_addr):
            raise RedisAdapterError('user mac bind error req={0}, bind={1}'.format(
                req_mac_addr, user.mac_address))

        # check vlanid1 bind
        if not RadiusUtils.check_in_vlan_bind(user, req.vlanid1):
            raise RedisAdapterError('user in vlanid bind error req={0}, bind={1}'.format(
                req.vlanid1, user.in_vlan))

        # check vlanid2 bind
        if not RadiusUtils.check_out_vlan_bind(user, req.vlanid2):
            raise RedisAdapterError('user outer vlanid bind error req={0}, bind={1}'.format(
                req.vlanid2, user.out_vlan))

        # check online limit
        cur_count = self.zrpc.radius_count_online(username,async=True).get()
        if cur_count >= user.online_limit:
            raise RedisAdapterError('current user online count %s > limit(%s) ' % (cur_count,user.online_limit))

        pre_reply = dict(code=0, msg='ok')
        pre_reply['ext_attrs']['input_rate'] = user.input_rate
        pre_reply['ext_attrs']['output_rate'] = user.output_rate
        pre_reply['radius_attrs']['Session-Timeout'] = RadiusUtils.calc_session_time(user.expire_time)

        for attrname, attrvalue in user.radius_attrs:
            pre_reply['ext_attrs'][attrname] = attrvalue

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
        _user = self.zrpc.radius_find_user(username)
        if not _user:
            raise RedisAdapterError('user {0} not exists or user status not enabled'.format(username))

        user = RadiusUser(_user)

        nasid = req.get_nas_id()
        sessionid = req.get_acct_sessionid()

        # check online limit
        cur_count = self.zrpc.radius_count_online(username,async=True).get()
        if cur_count >= user.online_limit:
            raise RedisAdapterError('current user online count %s > limit(%s) ' % (cur_count,user.online_limit))

        online = self.zrpc.radius_get_online(sessionid,username,async=True).get()
        if online:
            raise RedisAdapterError('user {0} session duplicate'.format(username))

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





