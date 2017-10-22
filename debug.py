#!/usr/bin/env python
#coding:utf-8
import gevent.monkey
gevent.monkey.patch_all()
import redis
from gevent import socket
import redis.connection
redis.connection.socket = socket
import os
import logging
import logging.config
import gevent
import signal
import sys
print sys.prefix

def run(conf):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from toughradius.radiusd.master import RudiusAuthServer
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        address = (config.radiusd.host, int(config.radiusd.auth_port))
        auth_server = RudiusAuthServer(address, config)


        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        from toughradius.radiusd.master import RudiusAcctServer
        address = (config.radiusd.host, int(config.radiusd.acct_port))
        acct_server = RudiusAcctServer(address, config)

        # gevent.signal(signal.SIGTERM, auth_server.close)
        # gevent.signal(signal.SIGINT, auth_server.close)
        # gevent.signal(signal.SIGTERM, acct_server.close)
        # gevent.signal(signal.SIGINT, acct_server.close)
        auth_server.start()
        gevent.sleep(0.1)
        acct_server.start()
        gevent.sleep(0.1)
        logging.info(auth_server)
        logging.info(acct_server)

        from toughradius.radiusd.apiserver import  ApiServer
        apiserver = ApiServer(config)
        apiserver.start(forever=False)
        gevent.wait()
    except:
        import traceback
        traceback.print_exc()




if __name__ == "__main__":
    run("etc/radiusd.json")
