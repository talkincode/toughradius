#!/usr/bin/env python
# -*- coding: utf-8 -*-
import os
import sys
sys.path.insert(0,os.path.dirname(__file__))
import re
import time
from toughradius.common import choosereactor
choosereactor.install_optimal_reactor(False)
import sys,signal,click
import platform as pf
from twisted.internet import reactor
from twisted.python import log
from toughradius.common import config as iconfig
from toughradius.common import dispatch,logger,utils
from toughradius.common.dbengine import get_engine
from toughradius.manage import settings
from toughradius.manage.settings import redis_conf
from toughradius.common import log_trace
import traceback

def setup_logger(config):
    syslog = logger.Logger(config,'radius')
    dispatch.register(syslog)
    log.startLoggingWithObserver(syslog.emit, setStdout=0)
    return syslog

def update_timezone(config):
    if 'TZ' not in os.environ:
        os.environ["TZ"] = config.system.tz
    try:time.tzset()
    except:pass

def reactor_run():
    def ExitHandler(signum, stackframe):
        print "Got signal: %s" % signum
        reactor.callFromThread(reactor.stop)
    signal.signal(signal.SIGTERM, ExitHandler)
    reactor.run()

def main():
    try:
        from toughradius import httpd
        from toughradius import radiusd
        from toughradius import taskd
        from toughradius.common.redis_cache import CacheManager
        config = iconfig.find_config('etc/toughradius.json')
        update_timezone(config)
        dbengine = get_engine(config)
        cache = CacheManager(redis_conf(config),cache_name='RadiusCache')
        aes = utils.AESCipher(key=config.system.secret)
        log = setup_logger(config)
        httpd.run(config,dbengine,cache=cache,aes=aes)
        radiusd.run_auth(config)
        radiusd.run_acct(config)
        radiusd.run_worker(config,dbengine,cache=cache,aes=aes,standalone=True)
        taskd.run(config,dbengine,cache=cache,aes=aes,standalone=True)
        reactor_run()        
    except:
        traceback.print_exc()

           
if __name__ == '__main__':
     main()
