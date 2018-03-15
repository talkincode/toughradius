#!/usr/bin/env python
# coding:utf-8
try:
    import zerorpc
except:
    pass
import gevent
from toughradius.common.mcache import cache
from .base import BasicAdapter
from toughradius.pyrad import message
import datetime
import logging
from toughradius.common import tools
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


class RedisAdapterError(Exception): pass

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
        gevent.spawn(self.init_rpc)

    def init_rpc(self):
        self.zrpc = zerorpc.Client(connect_to=self.settings.ADAPTERS['zerorpc']['connect'])

    def getClient(self, nasip=None, nasid=None):
        def fetch_result():
            return self.zrpc.radius_find_nas(nasip,nasid)
        return cache.aget('toughradius.nas.cache.{0}.{1}'.format(nasid, nasip), fetch_result, expire=60)

    def getUserAttr(self, username, attrname,default_val=None):
        def fetch_result():
            result =  self.zrpc.radius_get_userattr(username,attrname.get(timeout=3))
            if result is None:
                return default_val
        return cache.aget('toughradius.user.attr.cache.{0}.{1}'.format(username, attrname), fetch_result, expire=60)

    def processAuth(self, req):
        # check exists
        username = req.get_user_name()
        _user = self.zrpc.radius_find_user(username, async=True).get(timeout=3)
        if not _user:
            raise RedisAdapterError('user {0} not exists or user status not enabled'.format(username))

        user = RadiusUser(_user)

        # check  expire
        if RadiusUtils.is_expire(user.expire_time):
            raise RedisAdapterError('user is expire at {0}'.format(user.expire_time))

        # check password
        if int(self.settings.RADIUSD.get('ignore_password', 0)) == 0:
            if user.ignore_password == 0:
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
        cur_count = self.zrpc.radius_count_online(username,async=True).get(timeout=3)
        if cur_count >= user.online_limit:
            raise RedisAdapterError('current user online count %s > limit(%s) ' % (cur_count,user.online_limit))

        pre_reply = dict(code=0, msg='ok',radius_attrs={})
        pre_reply['input_rate'] = user.input_rate
        pre_reply['output_rate'] = user.output_rate
        pre_reply['radius_attrs']['Session-Timeout'] = RadiusUtils.calc_session_time(user.expire_time)

        for attrname, attrvalue in user.radius_attrs:
            pre_reply['radius_attrs'][attrname] = attrvalue

        return pre_reply

    def processAcct(self, req):
        status_type = req.get_acct_status_type()
        username = req.get_user_name()
        billing = req.get_billing()
        billing['nas_paddr'] = req.source[0]

        if status_type == message.STATUS_TYPE_START:
            self.zrpc.radius_start_accounting(billing, async=True)
            self.logger.info(u'user {0} start accounting'.format(tools.safeunicode(username)))
            return dict(code=0, msg='ok')
        elif status_type == message.STATUS_TYPE_UPDATE:
            self.zrpc.radius_update_accounting(billing, async=True)
            self.logger.info(u'user {0} update accounting'.format(tools.safeunicode(username)))
            return dict(code=0, msg='ok')
        elif status_type == message.STATUS_TYPE_STOP:
            self.zrpc.radius_stop_accounting(billing, async=True)
            self.logger.info(u'user {0} stop accounting'.format(tools.safeunicode(username)))
            return dict(code=0, msg='ok')
        elif status_type in (message.STATUS_TYPE_ACCT_ON, message.STATUS_TYPE_ACCT_OFF):
            nasid, nasip, nas_pub_addr = billing['nas_id'],billing['nas_addr'],billing['nas_paddr']
            self.zrpc.radius_onoff_accounting(nasid, nasip, nas_pub_addr, async=True)
            self.logger.info(u'on/off accounting nasid={0},nasip={1},nas_pub_addr={2}'.format(nasid, nasip, nas_pub_addr))
            return dict(code=0, msg='ok')


adapter = RedisAdapter





