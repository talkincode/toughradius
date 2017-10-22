#!/usr/bin/env python
#coding:utf-8
import gevent.monkey
gevent.monkey.patch_all()
from toughradius.common.bottle import Bottle,route, run, request
from toughradius.common.bottle_redis import RedisPlugin
from toughradius.common.ghttpd import GeventServer
from toughradius.common.rediskeys import UserAttrs, NasAttrs
from toughradius.common.rediskeys import UserStates, NasStates
from toughradius.common import rediskeys
import logging
import time

class ApiRoutes(object):

    def radtest(self, rdb):
        return dict(code=0,msg="success")

    def inittest(self,rdb):
        with rdb.pipeline() as pipe:
            user = dict(
                status=1,
                username='test01',
                password='888888',
                input_rate=86400,
                output_rate=86400,
                rate_code='',
                bill_type='day',
                bind_mac=0,
                bind_vlan=0,
                bind_nas=0,
                nas='',
                mac_addr='',
                vlanid1=0,
                vlanid2=0,
                time_amount=0,
                flow_amount=0,
                expire_date='2019-12-30',
                expire_time='23:59:59',
                online_limit=0,
                bypass_pwd=0
            )
            pipe.hmset(rediskeys.UserHKey('test01'),user)
            pipe.zadd(rediskeys.UserSetKey,time.time(), rediskeys.UserHKey('test01'))
            nas = dict(
                status=1,
                nasid='toughac',
                name='toughac',
                vendor=0,
                ipaddr='127.0.0.1',
                secret='secret',
                coaport=3799
            )
            pipe.hmset(rediskeys.NasHKey('toughac','127.0.0.1'),nas)
            pipe.zadd(rediskeys.NasSetKey,time.time(), rediskeys.NasHKey('toughac','127.0.0.1'))
            pipe.execute()
        return dict(code=0, msg="success")



class ApiServer(ApiRoutes):

    def __init__(self, config):
        self.config = config
        self.host = config.api['host']
        self.port=int(config.api['port'])
        self.app = Bottle()
        plugin = RedisPlugin(host='localhost')
        self.app.install(plugin)
        self.setup_routing()

    def setup_routing(self):
        self.app.route('/api/v1/radtest', ['GET', 'POST'], self.radtest)
        self.app.route('/api/v1/inittest', ['GET', 'POST'], self.inittest)

    def start(self, forever=True):
        return run(self.app, server=GeventServer, host=self.host, port=self.port,forever=forever)


if __name__ == '__main__':
    from toughradius.common import config as iconfig
    import logging.config
    config = iconfig.find_config("../../etc/radiusd.json")
    logging.config.dictConfig(config.logger)
    ApiServer(config).start(forever=True)

