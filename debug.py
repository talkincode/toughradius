#!/usr/bin/env python
#coding:utf-8
import gevent.monkey
gevent.monkey.patch_all()
import os
import logging
import logging.config
import gevent
import signal

def runauth(conf):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from toughradius.radiusd.master import RudiusAuthServer
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        address = (config.radiusd.host, int(config.radiusd.auth_port))
        server = RudiusAuthServer(address, config)
        server.start()
        logging.info(server)
        # server.serve_forever()
    except:
        import traceback
        traceback.print_exc()

def runacct(conf):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        from toughradius.radiusd.master import RudiusAcctServer
        address = (config.radiusd.host, int(config.radiusd.acct_port))
        server = RudiusAcctServer(address, config)
        server.start()
        logging.info(server)
        # server.serve_forever()
    except:
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    runauth("etc/radiusd.json")
    runacct("etc/radiusd.json")
    gevent.wait()